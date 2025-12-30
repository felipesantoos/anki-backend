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
	addonService "github.com/felipesantos/anki-backend/core/services/addon"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	backupService "github.com/felipesantos/anki-backend/core/services/backup"
	emailService "github.com/felipesantos/anki-backend/core/services/email"
	mediaService "github.com/felipesantos/anki-backend/core/services/media"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	syncService "github.com/felipesantos/anki-backend/core/services/sync"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	infraEmail "github.com/felipesantos/anki-backend/infra/email"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func TestSystem_Integration(t *testing.T) {
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
	addOnRepo := repositories.NewAddOnRepository(db.DB)
	backupRepo := repositories.NewBackupRepository(db.DB)
	mediaRepo := repositories.NewMediaRepository(db.DB)
	syncMetaRepo := repositories.NewSyncMetaRepository(db.DB)

	// Setup Services
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err = eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	emailRepo := infraEmail.NewConsoleRepository(log)
	emailSvc := emailService.NewEmailService(emailRepo, jwtSvc, cfg.Email)

	authSvc := authService.NewAuthService(userRepo, deckRepo, profileRepo, userPreferencesRepo, eventBus, jwtSvc, redisRepo, emailSvc, sessionSvc)
	addOnSvc := addonService.NewAddOnService(addOnRepo)
	backupSvc := backupService.NewBackupService(backupRepo)
	mediaSvc := mediaService.NewMediaService(mediaRepo)
	syncMetaSvc := syncService.NewSyncMetaService(syncMetaRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc, jwtSvc, redisRepo, sessionSvc)
	routes.RegisterSystemRoutes(e, addOnSvc, backupSvc, mediaSvc, syncMetaSvc, jwtSvc, redisRepo)

	// Register and login
	loginRes := registerAndLogin(t, e, "system@example.com", "password123")
	token := loginRes.AccessToken

	t.Run("AddOns", func(t *testing.T) {
		// Install AddOn
		installReq := request.InstallAddOnRequest{
			Code:       "12345",
			Name:       "Test AddOn",
			Version:    "1.0.0",
			ConfigJSON: "{}",
		}
		b, _ := json.Marshal(installReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/addons", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var addonRes response.AddOnResponse
		json.Unmarshal(rec.Body.Bytes(), &addonRes)
		assert.Equal(t, installReq.Code, addonRes.Code)

		// List AddOns
		req = httptest.NewRequest(http.MethodGet, "/api/v1/addons", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var addonsRes []response.AddOnResponse
		json.Unmarshal(rec.Body.Bytes(), &addonsRes)
		assert.NotEmpty(t, addonsRes)

		// Update Config
		updateReq := request.UpdateAddOnConfigRequest{
			ConfigJSON: `{"theme": "dark"}`,
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/addons/12345/config", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Toggle AddOn
		toggleReq := request.ToggleAddOnRequest{Enabled: false}
		b, _ = json.Marshal(toggleReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/addons/12345/toggle", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Uninstall AddOn
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/addons/12345", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Backups", func(t *testing.T) {
		// Create Backup
		createReq := request.CreateBackupRequest{
			Filename:    "backup_test.colpkg",
			BackupType:  "manual",
			Size:        1024,
			StoragePath: "/storage/backup_test.colpkg",
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/backups", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var backupRes response.BackupResponse
		json.Unmarshal(rec.Body.Bytes(), &backupRes)
		assert.Equal(t, createReq.Filename, backupRes.Filename)
		backupID := backupRes.ID

		// List Backups
		req = httptest.NewRequest(http.MethodGet, "/api/v1/backups", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var backupsRes []response.BackupResponse
		json.Unmarshal(rec.Body.Bytes(), &backupsRes)
		assert.NotEmpty(t, backupsRes)

		// Delete Backup
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/backups/"+strconv.FormatInt(backupID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Media", func(t *testing.T) {
		// Create Media
		createReq := request.CreateMediaRequest{
			Filename:    "test_image.jpg",
			Hash:        "abc123hash",
			Size:        1024,
			MimeType:    "image/jpeg",
			StoragePath: "/storage/test_image.jpg",
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/media", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var mediaRes response.MediaResponse
		json.Unmarshal(rec.Body.Bytes(), &mediaRes)
		assert.Equal(t, createReq.Filename, mediaRes.Filename)
		mediaID := mediaRes.ID

		// Find All Media
		req = httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var mediasRes []response.MediaResponse
		json.Unmarshal(rec.Body.Bytes(), &mediasRes)
		assert.NotEmpty(t, mediasRes)

		// Find Media by ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/media/"+strconv.FormatInt(mediaID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &mediaRes)
		assert.Equal(t, createReq.Filename, mediaRes.Filename)

		// Delete Media
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+strconv.FormatInt(mediaID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("SyncMeta", func(t *testing.T) {
		// Get SyncMeta (initially it might not exist or be empty, but service should handle it)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/meta", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Should return OK with data
		assert.Equal(t, http.StatusOK, rec.Code)

		// Update SyncMeta
		updateReq := request.UpdateSyncMetaRequest{
			ClientID:    "mobile-app-1",
			LastSyncUSN: 100,
		}
		b, _ := json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/sync/meta", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var syncMetaRes response.SyncMetaResponse
		json.Unmarshal(rec.Body.Bytes(), &syncMetaRes)
		assert.Equal(t, updateReq.ClientID, syncMetaRes.ClientID)
		assert.Equal(t, updateReq.LastSyncUSN, syncMetaRes.LastSyncUSN)
	})
}

