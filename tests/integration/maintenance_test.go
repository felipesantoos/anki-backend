package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/dicontainer"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func TestMaintenance_Integration(t *testing.T) {
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
	router.RegisterContentRoutes()
	router.RegisterMaintenanceRoutes()

	// Register and login
	loginRes := registerAndLogin(t, e, "maintenance@example.com", "password123")
	token := loginRes.AccessToken

	// 1. Setup: Create NoteType with conditional template
	createNTReq := request.CreateNoteTypeRequest{
		Name:          "Conditional Note Type",
		FieldsJSON:    `[{"name": "Front"}, {"name": "Back"}, {"name": "Extra"}]`,
		CardTypesJSON: `[{"name": "Card 1"}, {"name": "Card 2"}]`,
		TemplatesJSON: `[{"qfmt":"{{Front}}","afmt":"{{Back}}"},{"qfmt":"{{#Extra}}{{Extra}}{{/Extra}}","afmt":"{{Back}}"}]`,
	}
	b, _ := json.Marshal(createNTReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/note-types", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var ntRes response.NoteTypeResponse
	json.Unmarshal(rec.Body.Bytes(), &ntRes)
	ntID := ntRes.ID

	// 2. Get default deck
	req = httptest.NewRequest(http.MethodGet, "/api/v1/decks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var decksRes []response.DeckResponse
	json.Unmarshal(rec.Body.Bytes(), &decksRes)
	deckID := decksRes[0].ID

	// 3. Create Note with "Extra" field (generates 2 cards)
	createNoteReq := request.CreateNoteRequest{
		NoteTypeID: ntID,
		DeckID:     deckID,
		FieldsJSON: `{"Front":"Q1","Back":"A1","Extra":"Something"}`,
		Tags:       []string{},
	}
	b, _ = json.Marshal(createNoteReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// 4. Change NoteType template so Card 2 becomes empty even WITH "Extra"
	updateNTReq := request.UpdateNoteTypeRequest{
		Name:          "Conditional Note Type",
		FieldsJSON:    `[{"name": "Front"}, {"name": "Back"}, {"name": "Extra"}]`,
		CardTypesJSON: `[{"name": "Card 1"}, {"name": "Card 2"}]`,
		TemplatesJSON: `[{"qfmt":"{{Front}}","afmt":"{{Back}}"},{"qfmt":"{{#MissingField}}{{Extra}}{{/MissingField}}","afmt":"{{Back}}"}]`,
	}
	b, _ = json.Marshal(updateNTReq)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/note-types/%d", ntID), bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Now Card 2 of the previously created note is "empty" because it refers to MissingField.
	
	// 5. Test GetEmptyCards
	req = httptest.NewRequest(http.MethodGet, "/api/v1/maintenance/empty-cards", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var emptyRes response.EmptyCardsResponse
	json.Unmarshal(rec.Body.Bytes(), &emptyRes)
	assert.True(t, emptyRes.Count >= 1, "Should find at least one empty card")

	// 6. Test CleanupEmptyCards
	req = httptest.NewRequest(http.MethodPost, "/api/v1/maintenance/empty-cards/cleanup", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var cleanupRes response.CleanupEmptyCardsResponse
	json.Unmarshal(rec.Body.Bytes(), &cleanupRes)
	assert.True(t, cleanupRes.DeletedCount >= 1, "Should have deleted at least one card")

	// 7. Verify no more empty cards
	req = httptest.NewRequest(http.MethodGet, "/api/v1/maintenance/empty-cards", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &emptyRes)
	assert.Equal(t, 0, emptyRes.Count, "Should find zero empty cards after cleanup")
}
