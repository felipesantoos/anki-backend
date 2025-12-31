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

	// Disable rate limiting for tests
	cfg.RateLimit.Enabled = false

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
	router.Init()

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

		// Delete Deck
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/decks/"+strconv.FormatInt(deckID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify Deck is gone
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+strconv.FormatInt(deckID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("DeckOptions", func(t *testing.T) {
		// 1. Create Deck with specific options
		options := map[string]interface{}{
			"newCardsPerDay": 25,
			"reviewLimit":    100,
		}
		optionsJSON, _ := json.Marshal(options)
		createReq := request.CreateDeckRequest{
			Name:        "Options Test Deck",
			OptionsJSON: string(optionsJSON),
		}
		d := createDeck(t, e, token, createReq)

		// 2. Get options via endpoint
		req := httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+strconv.FormatInt(d.ID, 10)+"/options", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var retrievedOptions map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &retrievedOptions)

		assert.Equal(t, float64(25), retrievedOptions["newCardsPerDay"])
		assert.Equal(t, float64(100), retrievedOptions["reviewLimit"])

		// 3. Update options
		updatedOptions := map[string]interface{}{
			"newCardsPerDay": 30,
			"reviewLimit":    150,
			"newOption":      true,
		}
		b, _ := json.Marshal(updatedOptions)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(d.ID, 10)+"/options", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &retrievedOptions)
		assert.Equal(t, float64(30), retrievedOptions["newCardsPerDay"])
		assert.Equal(t, float64(150), retrievedOptions["reviewLimit"])
		assert.Equal(t, true, retrievedOptions["newOption"])

		// 4. Try to get options for non-existent deck
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/999999/options", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("DeckStats", func(t *testing.T) {
		// Create a deck
		deckReq := request.CreateDeckRequest{Name: "Stats Test Deck"}
		deckRes := createDeck(t, e, token, deckReq)
		deckID := deckRes.ID

		// Get Stats (should be empty but exist)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+strconv.FormatInt(deckID, 10)+"/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var stats response.DeckStatsResponse
		json.Unmarshal(rec.Body.Bytes(), &stats)
		assert.Equal(t, deckID, stats.DeckID)
		assert.Equal(t, 0, stats.NewCount)
		assert.Equal(t, 0, stats.NotesCount)

		// Verification of stats with data is better suited for a larger integration test
		// but this confirms the endpoint is working and correctly mapped.
	})

	t.Run("DeckOptionsPresets", func(t *testing.T) {
		// 1. Create Preset
		createReq := request.CreateDeckOptionsPresetRequest{
			Name:        "Custom Preset",
			OptionsJSON: `{"newCardsPerDay": 50}`,
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/deck-options-presets", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var presetRes response.DeckOptionsPresetResponse
		json.Unmarshal(rec.Body.Bytes(), &presetRes)
		assert.Equal(t, createReq.Name, presetRes.Name)
		assert.Equal(t, createReq.OptionsJSON, presetRes.OptionsJSON)
		presetID := presetRes.ID

		// 2. List Presets
		req = httptest.NewRequest(http.MethodGet, "/api/v1/deck-options-presets", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var presetList []response.DeckOptionsPresetResponse
		json.Unmarshal(rec.Body.Bytes(), &presetList)
		assert.NotEmpty(t, presetList)

		// 3. Update Preset
		updateReq := request.UpdateDeckOptionsPresetRequest{
			Name:        "Updated Preset",
			OptionsJSON: `{"newCardsPerDay": 60}`,
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/deck-options-presets/"+strconv.FormatInt(presetID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &presetRes)
		assert.Equal(t, updateReq.Name, presetRes.Name)

		// 4. Delete Preset
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/deck-options-presets/"+strconv.FormatInt(presetID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// 5. Apply to Decks
		// Create a new preset
		createReq = request.CreateDeckOptionsPresetRequest{
			Name:        "Apply Test Preset",
			OptionsJSON: `{"newCardsPerDay": 100}`,
		}
		b, _ = json.Marshal(createReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/deck-options-presets", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &presetRes)
		newPresetID := presetRes.ID

		// Create two decks
		deck1 := createDeck(t, e, token, request.CreateDeckRequest{Name: "Deck 1"})
		deck2 := createDeck(t, e, token, request.CreateDeckRequest{Name: "Deck 2"})

		// Apply preset
		applyReq := request.ApplyDeckOptionsPresetRequest{
			DeckIDs: []int64{deck1.ID, deck2.ID},
		}
		b, _ = json.Marshal(applyReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/deck-options-presets/"+strconv.FormatInt(newPresetID, 10)+"/apply", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify decks options
		for _, did := range applyReq.DeckIDs {
			req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+strconv.FormatInt(did, 10), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec = httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			var dRes response.DeckResponse
			json.Unmarshal(rec.Body.Bytes(), &dRes)
			assert.Equal(t, `{"newCardsPerDay": 100}`, dRes.OptionsJSON)
		}
	})

	t.Run("DeckHierarchy", func(t *testing.T) {
		// Create Parent
		parentReq := request.CreateDeckRequest{Name: "Parent"}
		b, _ := json.Marshal(parentReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var parentRes response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &parentRes)

		// Create Child
		childReq := request.CreateDeckRequest{
			Name:     "Child",
			ParentID: &parentRes.ID,
		}
		b, _ = json.Marshal(childReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var childRes response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &childRes)

		// Verify FullName in List
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var deckList []response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &deckList)

		foundChild := false
		for _, d := range deckList {
			if d.ID == childRes.ID {
				assert.Equal(t, "Parent::Child", d.FullName)
				foundChild = true
			}
		}
		assert.True(t, foundChild)
	})

	t.Run("Reorganization", func(t *testing.T) {
		// 1. Create structure: Parent A, Parent B, Child A (under Parent A)
		parentAReq := request.CreateDeckRequest{Name: "Parent A"}
		pA := createDeck(t, e, token, parentAReq)

		parentBReq := request.CreateDeckRequest{Name: "Parent B"}
		pB := createDeck(t, e, token, parentBReq)

		childAReq := request.CreateDeckRequest{Name: "Child A", ParentID: &pA.ID}
		cA := createDeck(t, e, token, childAReq)

		// 2. Move Child A to Parent B
		moveReq := request.UpdateDeckRequest{Name: "Child A", ParentID: &pB.ID}
		b, _ := json.Marshal(moveReq)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(cA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var movedRes response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &movedRes)
		assert.Equal(t, &pB.ID, movedRes.ParentID)
		
		// Verify FullName updated in list
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		var movedResList []response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &movedResList)
		found := false
		for _, d := range movedResList {
			if d.ID == cA.ID {
				assert.Equal(t, "Parent B::Child A", d.FullName)
				found = true
			}
		}
		assert.True(t, found)

		// 3. Move Child A to root
		moveRootReq := request.UpdateDeckRequest{Name: "Child A", ParentID: nil}
		b, _ = json.Marshal(moveRootReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(cA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &movedRes)
		assert.Nil(t, movedRes.ParentID)

		// 4. Try to move Parent A into itself (should fail)
		moveSelfReq := request.UpdateDeckRequest{Name: "Parent A", ParentID: &pA.ID}
		b, _ = json.Marshal(moveSelfReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(pA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// 5. Deep Circular Dependency
		// Move Child A back to Parent B
		moveReq.ParentID = &pB.ID
		b, _ = json.Marshal(moveReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(cA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		// Now try to move Parent B into Child A
		moveInvalidReq := request.UpdateDeckRequest{Name: "Parent B", ParentID: &cA.ID}
		b, _ = json.Marshal(moveInvalidReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(pB.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Cards", func(t *testing.T) {
		// Create a new deck for card operations
		createReq := request.CreateDeckRequest{
			Name: "Card Test Deck",
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var deckRes response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &deckRes)
		cardDeckID := deckRes.ID

		// Manually create a note and card for testing card operations
		var noteTypeID int64
		err := db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'Test Type', '[]', '[]', '[]') RETURNING id", loginRes.User.ID).Scan(&noteTypeID)
		require.NoError(t, err)

		var noteID int64
		// Use a valid UUID for guid
		err = db.DB.QueryRow("INSERT INTO notes (user_id, note_type_id, fields_json, guid) VALUES ($1, $2, '{}', '550e8400-e29b-41d4-a716-446655440000') RETURNING id", loginRes.User.ID, noteTypeID).Scan(&noteID)
		require.NoError(t, err)

		var cardID int64
		err = db.DB.QueryRow("INSERT INTO cards (deck_id, note_id, card_type_id, state) VALUES ($1, $2, 0, 'new') RETURNING id", cardDeckID, noteID).Scan(&cardID)
		require.NoError(t, err)

		// Get Card
		req = httptest.NewRequest(http.MethodGet, "/api/v1/cards/"+strconv.FormatInt(cardID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
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
		b = []byte{} // reset b
		b, err = json.Marshal(flagReq)
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

// Helper function to create a deck and return its response
func createDeck(t *testing.T, e *echo.Echo, token string, reqBody request.CreateDeckRequest) response.DeckResponse {
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var deckRes response.DeckResponse
	json.Unmarshal(rec.Body.Bytes(), &deckRes)
	return deckRes
}

