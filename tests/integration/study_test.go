package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	"github.com/felipesantos/anki-backend/dicontainer"
	"github.com/felipesantos/anki-backend/config"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
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

	// Setup Services
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err = eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// Initialize DI Package
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()
	router.RegisterStudyRoutes()

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

		// List Decks
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var deckList []response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &deckList)
		assert.NotEmpty(t, deckList)
		
		// Find our created deck in the list
		found := false
		for _, d := range deckList {
			if d.ID == deckID {
				found = true
				assert.Equal(t, "Updated Integration Deck", d.Name)
				break
			}
		}
		assert.True(t, found, "Created deck should be in the list")
	})

	t.Run("Cards", func(t *testing.T) {
		// Manually create a note and card for testing card operations
		var noteTypeID int64
		err := db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'Test Type', '[]', '[]', '[]') RETURNING id", loginRes.User.ID).Scan(&noteTypeID)
		require.NoError(t, err)

		var noteID int64
		// Use a valid UUID for guid
		err = db.DB.QueryRow("INSERT INTO notes (user_id, note_type_id, fields_json, guid) VALUES ($1, $2, '{}', '550e8400-e29b-41d4-a716-446655440000') RETURNING id", loginRes.User.ID, noteTypeID).Scan(&noteID)
		require.NoError(t, err)

		var cardID int64
		err = db.DB.QueryRow("INSERT INTO cards (deck_id, note_id, card_type_id, state) VALUES ($1, $2, 0, 'new') RETURNING id", deckID, noteID).Scan(&cardID)
		require.NoError(t, err)

		// Get Card
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/"+strconv.FormatInt(cardID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var cardRes response.CardResponse
		json.Unmarshal(rec.Body.Bytes(), &cardRes)
		assert.Equal(t, cardID, cardRes.ID)

		// Suspend Card
		req = httptest.NewRequest(http.MethodPost, "/api/v1/cards/"+strconv.FormatInt(cardID, 10)+"/suspend", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Unsuspend Card
		req = httptest.NewRequest(http.MethodPost, "/api/v1/cards/"+strconv.FormatInt(cardID, 10)+"/unsuspend", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Bury Card
		req = httptest.NewRequest(http.MethodPost, "/api/v1/cards/"+strconv.FormatInt(cardID, 10)+"/bury", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Unbury Card
		req = httptest.NewRequest(http.MethodPost, "/api/v1/cards/"+strconv.FormatInt(cardID, 10)+"/unbury", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Set Flag
		flagReq := request.SetCardFlagRequest{Flag: 1}
		b, err := json.Marshal(flagReq)
		require.NoError(t, err)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/cards/"+strconv.FormatInt(cardID, 10)+"/flag", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Reviews", func(t *testing.T) {
		// We need a card ID from previous test or create a new one
		var cardID int64
		err := db.DB.QueryRow("SELECT id FROM cards LIMIT 1").Scan(&cardID)
		require.NoError(t, err)

		// Create Review
		createReq := request.CreateReviewRequest{
			CardID: cardID,
			Rating: 3,
			TimeMs: 10000,
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var reviewRes response.ReviewResponse
		json.Unmarshal(rec.Body.Bytes(), &reviewRes)
		assert.Equal(t, createReq.Rating, reviewRes.Rating)

		// Find Reviews by Card ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/cards/"+strconv.FormatInt(cardID, 10)+"/reviews", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var reviewsRes []response.ReviewResponse
		json.Unmarshal(rec.Body.Bytes(), &reviewsRes)
		assert.NotEmpty(t, reviewsRes)
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

