package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	emailService "github.com/felipesantos/anki-backend/core/services/email"
	profileService "github.com/felipesantos/anki-backend/core/services/profile"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	userService "github.com/felipesantos/anki-backend/core/services/user"
	userpreferencesService "github.com/felipesantos/anki-backend/core/services/userpreferences"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	infraEmail "github.com/felipesantos/anki-backend/infra/email"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func TestUser_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := logger.GetLogger()
	cfg, err := config.Load()
	require.NoError(t, err)

	// Setup Redis
	redisRepo, err := redisInfra.NewRedisRepository(cfg.Redis, log)
	require.NoError(t, err)
	defer redisRepo.Close()

	// Setup JWT
	jwtSvc, err := jwt.NewJWTService(cfg.JWT)
	require.NoError(t, err)

	// Setup Session
	sessionRepo := redisInfra.NewSessionRepository(redisRepo.Client, cfg.Session.KeyPrefix)
	sessionTTL := time.Duration(cfg.Session.TTLMinutes) * time.Minute
	sessionSvc := sessionService.NewSessionService(sessionRepo, sessionTTL)

	// Setup Repositories
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)
	userPreferencesRepo := repositories.NewUserPreferencesRepository(db.DB)

	// Setup Services
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err = eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	emailRepo := infraEmail.NewConsoleRepository(log)
	emailSvc := emailService.NewEmailService(emailRepo, jwtSvc, cfg.Email)

	authSvc := authService.NewAuthService(userRepo, deckRepo, profileRepo, userPreferencesRepo, eventBus, jwtSvc, redisRepo, emailSvc, sessionSvc)
	userSvc := userService.NewUserService(userRepo)
	profileSvc := profileService.NewProfileService(profileRepo)
	userPreferencesSvc := userpreferencesService.NewUserPreferencesService(userPreferencesRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc, jwtSvc, redisRepo, sessionSvc)
	routes.RegisterUserRoutes(e, userSvc, profileSvc, userPreferencesSvc, jwtSvc, redisRepo)

	// Register and login
	loginRes := registerAndLogin(t, e, "user@example.com", "password123")
	token := loginRes.AccessToken

	t.Run("User", func(t *testing.T) {
		// Get Me
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var userRes response.UserResponse
		json.Unmarshal(rec.Body.Bytes(), &userRes)
		assert.Equal(t, "user@example.com", userRes.Email)

		// Update Email
		updateReq := request.UpdateUserRequest{
			Email: "new-email@example.com",
		}
		b, _ := json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/user/me", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &userRes)
		assert.Equal(t, updateReq.Email, userRes.Email)
	})

	t.Run("Profiles", func(t *testing.T) {
		// List Profiles (should have at least the default created during registration)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var profilesRes []response.ProfileResponse
		json.Unmarshal(rec.Body.Bytes(), &profilesRes)
		assert.GreaterOrEqual(t, len(profilesRes), 1)

		// Create Profile
		createReq := request.CreateProfileRequest{
			Name: "Secondary Profile",
		}
		b, _ := json.Marshal(createReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var profileRes response.ProfileResponse
		json.Unmarshal(rec.Body.Bytes(), &profileRes)
		assert.Equal(t, createReq.Name, profileRes.Name)
		profileID := profileRes.ID

		// Update Profile
		updateReq := request.UpdateProfileRequest{
			Name: "Renamed Profile",
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/profiles/"+strconv.FormatInt(profileID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &profileRes)
		assert.Equal(t, updateReq.Name, profileRes.Name)

		// Enable Sync
		syncReq := request.EnableSyncRequest{Username: "testuser"}
		b, _ = json.Marshal(syncReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles/"+strconv.FormatInt(profileID, 10)+"/sync/enable", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Disable Sync
		req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles/"+strconv.FormatInt(profileID, 10)+"/sync/disable", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Delete Profile
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/profiles/"+strconv.FormatInt(profileID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("UserPreferences", func(t *testing.T) {
		// Get Preferences
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var prefsRes response.UserPreferencesResponse
		json.Unmarshal(rec.Body.Bytes(), &prefsRes)

		// Update Preferences
		updateReq := request.UpdateUserPreferencesRequest{
			Language:               "pt-BR",
			Theme:                  "dark",
			UISize:                 1.2,
			AutoSync:               true,
			DefaultDeckBehavior:    "current_deck",
			NextDayStartsAt:        time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC),
			LearnAheadLimit:        20,
			TimeboxTimeLimit:       0,
			VideoDriver:            "auto",
			ShowPlayButtons:        true,
			InterruptAudioOnAnswer: true,
			ShowRemainingCount:     true,
			SpacebarAnswersCard:    true,
			SyncAudioAndImages:     true,
		}
		b, _ := json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/user/preferences", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &prefsRes)
		assert.Equal(t, updateReq.Language, prefsRes.Language)
		assert.Equal(t, updateReq.Theme, prefsRes.Theme)

		// Reset to Defaults
		req = httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences/reset", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("UserDelete", func(t *testing.T) {
		// Delete User (Soft Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify user is soft deleted (should not be able to get me anymore)
		req = httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

