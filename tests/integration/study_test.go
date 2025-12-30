package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	cardService "github.com/felipesantos/anki-backend/core/services/card"
	deckService "github.com/felipesantos/anki-backend/core/services/deck"
	emailService "github.com/felipesantos/anki-backend/core/services/email"
	reviewService "github.com/felipesantos/anki-backend/core/services/review"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	infraEmail "github.com/felipesantos/anki-backend/infra/email"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/database"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func TestStudy_Integration(t *testing.T) {
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
	tm := database.NewTransactionManager(db.DB)
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)
	userPreferencesRepo := repositories.NewUserPreferencesRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)
	reviewRepo := repositories.NewReviewRepository(db.DB)
	filteredDeckRepo := repositories.NewFilteredDeckRepository(db.DB)

	// Setup Services
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err = eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	emailRepo := infraEmail.NewConsoleRepository(log)
	emailSvc := emailService.NewEmailService(emailRepo, jwtSvc, cfg.Email)

	authSvc := authService.NewAuthService(userRepo, deckRepo, profileRepo, userPreferencesRepo, eventBus, jwtSvc, redisRepo, emailSvc, sessionSvc)
	deckSvc := deckService.NewDeckService(deckRepo)
	cardSvc := cardService.NewCardService(cardRepo)
	reviewSvc := reviewService.NewReviewService(reviewRepo, cardRepo, tm)
	filteredDeckSvc := deckService.NewFilteredDeckService(filteredDeckRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc, jwtSvc, redisRepo, sessionSvc)
	routes.RegisterStudyRoutes(e, deckSvc, filteredDeckSvc, cardSvc, reviewSvc, jwtSvc, redisRepo)

	// Register and login
	loginRes := registerAndLogin(t, e, "study@example.com", "password123")
	token := loginRes.AccessToken

	var deckID int64

	t.Run("Decks", func(t *testing.T) {
		// Create Deck
		createReq := request.CreateDeckRequest{
			Name: "Integration Test Deck",
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var deckRes response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &deckRes)
		assert.Equal(t, createReq.Name, deckRes.Name)
		deckID = deckRes.ID

		// Update Deck
		updateReq := request.UpdateDeckRequest{
			Name: "Updated Integration Deck",
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(deckID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &deckRes)
		assert.Equal(t, updateReq.Name, deckRes.Name)

		// Get Deck
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+strconv.FormatInt(deckID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &deckRes)
		assert.Equal(t, "Updated Integration Deck", deckRes.Name)
	})

	t.Run("Cards", func(t *testing.T) {
		// Since we haven't created notes yet, we'll just check if we can list cards for a deck (should be empty)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/decks/%d/cards", deckID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var cardsRes []response.CardResponse
		json.Unmarshal(rec.Body.Bytes(), &cardsRes)
		assert.Len(t, cardsRes, 0)
	})

	t.Run("FilteredDecks", func(t *testing.T) {
		// Create Filtered Deck
		createReq := request.CreateFilteredDeckRequest{
			Name:         "Integration Filtered Deck",
			SearchFilter: "is:due",
			Limit:        100,
			OrderBy:      "random",
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/filtered-decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var fdRes response.FilteredDeckResponse
		json.Unmarshal(rec.Body.Bytes(), &fdRes)
		assert.Equal(t, createReq.Name, fdRes.Name)
		fdID := fdRes.ID

		// Update Filtered Deck
		updateReq := request.UpdateFilteredDeckRequest{
			Name:         "Updated Filtered",
			SearchFilter: "tag:test",
			Limit:        50,
			OrderBy:      "newest",
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/filtered-decks/"+strconv.FormatInt(fdID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &fdRes)
		assert.Equal(t, updateReq.Name, fdRes.Name)
	})
}

