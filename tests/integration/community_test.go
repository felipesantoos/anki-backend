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
	auditService "github.com/felipesantos/anki-backend/core/services/audit"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	emailService "github.com/felipesantos/anki-backend/core/services/email"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	shareddeckService "github.com/felipesantos/anki-backend/core/services/shareddeck"
	shareddeckratingService "github.com/felipesantos/anki-backend/core/services/shareddeckrating"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	infraEmail "github.com/felipesantos/anki-backend/infra/email"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func strPtr(s string) *string {
	return &s
}

func TestCommunity_Integration(t *testing.T) {
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
	sharedDeckRepo := repositories.NewSharedDeckRepository(db.DB)
	sharedDeckRatingRepo := repositories.NewSharedDeckRatingRepository(db.DB)
	deletionLogRepo := repositories.NewDeletionLogRepository(db.DB)
	undoHistoryRepo := repositories.NewUndoHistoryRepository(db.DB)

	// Setup Services
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err = eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	emailRepo := infraEmail.NewConsoleRepository(log)
	emailSvc := emailService.NewEmailService(emailRepo, jwtSvc, cfg.Email)

	authSvc := authService.NewAuthService(userRepo, deckRepo, profileRepo, userPreferencesRepo, eventBus, jwtSvc, redisRepo, emailSvc, sessionSvc)
	sharedDeckSvc := shareddeckService.NewSharedDeckService(sharedDeckRepo)
	sharedDeckRatingSvc := shareddeckratingService.NewSharedDeckRatingService(sharedDeckRatingRepo)
	deletionLogSvc := auditService.NewDeletionLogService(deletionLogRepo)
	undoHistorySvc := auditService.NewUndoHistoryService(undoHistoryRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc, jwtSvc, redisRepo, sessionSvc)
	routes.RegisterCommunityRoutes(e, sharedDeckSvc, sharedDeckRatingSvc, deletionLogSvc, undoHistorySvc, jwtSvc, redisRepo)

	// Register and login
	loginRes := registerAndLogin(t, e, "community@example.com", "password123")
	token := loginRes.AccessToken

	var sharedDeckID int64

	t.Run("Marketplace", func(t *testing.T) {
		// First create a real deck to share
		var deckID int64
		err := db.DB.QueryRow("SELECT id FROM decks WHERE user_id = $1 AND name = 'Default'", loginRes.User.ID).Scan(&deckID)
		require.NoError(t, err)

		// Create SharedDeck
		createReq := request.CreateSharedDeckRequest{
			Name:        "Community Test Shared Deck",
			Description: strPtr("A deck for integration testing"),
			PackagePath: "/storage/test.colpkg",
			PackageSize: 1024,
			Tags:        []string{"test", "integration"},
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var sdRes response.SharedDeckResponse
		json.Unmarshal(rec.Body.Bytes(), &sdRes)
		assert.Equal(t, createReq.Name, sdRes.Name)
		sharedDeckID = sdRes.ID

		// List SharedDecks (Public)
		req = httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/decks", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var sdsRes []response.SharedDeckResponse
		json.Unmarshal(rec.Body.Bytes(), &sdsRes)
		assert.NotEmpty(t, sdsRes)

		// Find SharedDeck by ID (Public)
		req = httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/decks/"+strconv.FormatInt(sharedDeckID, 10), nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &sdRes)
		assert.Equal(t, "Community Test Shared Deck", sdRes.Name)

		// Update SharedDeck
		updateReq := request.UpdateSharedDeckRequest{
			Name:        "Updated Community Title",
			Description: strPtr("Updated description"),
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/marketplace/decks/"+strconv.FormatInt(sharedDeckID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Download SharedDeck
		req = httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/decks/"+strconv.FormatInt(sharedDeckID, 10)+"/download", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Ratings", func(t *testing.T) {
		// Create Rating
		createReq := request.CreateSharedDeckRatingRequest{
			SharedDeckID: sharedDeckID,
			Rating:       5,
			Comment:      strPtr("Great deck!"),
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/ratings", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var ratingRes response.SharedDeckRatingResponse
		json.Unmarshal(rec.Body.Bytes(), &ratingRes)
		assert.Equal(t, createReq.Rating, ratingRes.Rating)

		// List Ratings for Deck (Public)
		req = httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/decks/"+strconv.FormatInt(sharedDeckID, 10)+"/ratings", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var ratingsRes []response.SharedDeckRatingResponse
		json.Unmarshal(rec.Body.Bytes(), &ratingsRes)
		assert.NotEmpty(t, ratingsRes)

		// Update Rating
		updateReq := request.UpdateSharedDeckRatingRequest{
			Rating:  4,
			Comment: strPtr("Actually, it's a 4."),
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/marketplace/decks/"+strconv.FormatInt(sharedDeckID, 10)+"/ratings", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Delete Rating
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/marketplace/decks/"+strconv.FormatInt(sharedDeckID, 10)+"/ratings", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Audit", func(t *testing.T) {
		// Get Deletion Logs
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/deletions", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Get Undo History
		req = httptest.NewRequest(http.MethodGet, "/api/v1/audit/undo", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

