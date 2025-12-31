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

func TestContent_Integration(t *testing.T) {
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
	router.RegisterContentRoutes()

	// Register and login
	loginRes := registerAndLogin(t, e, "content@example.com", "password123")
	token := loginRes.AccessToken

	var noteTypeID int64

	t.Run("NoteTypes", func(t *testing.T) {
		// Create NoteType
		createReq := request.CreateNoteTypeRequest{
			Name:            "Basic Integration",
			FieldsJSON:      `[{"name": "Front"}, {"name": "Back"}]`,
			CardTypesJSON:   `[{"name": "Card 1"}]`,
			TemplatesJSON:   `[{"name": "Template 1"}]`,
		}
		b, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/note-types", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		assert.Equal(t, http.StatusCreated, rec.Code)
		var ntRes response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &ntRes)
		assert.Equal(t, createReq.Name, ntRes.Name)
		noteTypeID = ntRes.ID

		// Update NoteType
		updateReq := request.UpdateNoteTypeRequest{
			Name:          "Updated Basic Integration",
			FieldsJSON:     `[{"name": "Front"}, {"name": "Back"}, {"name": "Extra"}]`,
			CardTypesJSON:  `[{"name": "Card 1"}]`,
			TemplatesJSON:  `[{"name": "Template 1"}]`,
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/note-types/"+strconv.FormatInt(noteTypeID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &ntRes)
		assert.Equal(t, updateReq.Name, ntRes.Name)

		// List NoteTypes
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var ntListRes []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &ntListRes)
		assert.NotEmpty(t, ntListRes)

		// Find NoteType by ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types/"+strconv.FormatInt(noteTypeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Delete NoteType
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/note-types/"+strconv.FormatInt(noteTypeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Notes", func(t *testing.T) {
		// Re-create a note type for notes test since we deleted it above
		var noteTypeID int64
		err := db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'For Notes Test', '[]', '[{\"name\": \"Card 1\"}]', '[]') RETURNING id", loginRes.User.ID).Scan(&noteTypeID)
		require.NoError(t, err)

		// Get default deck ID
		var defaultDeckID int64
		err = db.DB.QueryRow("SELECT id FROM decks WHERE user_id = $1 AND name = 'Default'", loginRes.User.ID).Scan(&defaultDeckID)
		require.NoError(t, err)

		// Create Note
		createNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Integration Front", "Back": "Integration Back"}`,
			Tags:       []string{"integration", "test"},
		}
		b, _ := json.Marshal(createNoteReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		assert.Equal(t, http.StatusCreated, rec.Code)
		var noteRes response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		assert.Equal(t, createNoteReq.FieldsJSON, noteRes.FieldsJSON)
		noteID := noteRes.ID

		// Update Note
		updateReq := request.UpdateNoteRequest{
			FieldsJSON: `{"Front": "Updated Front", "Back": "Updated Back"}`,
			Tags:       []string{"integration", "updated"},
		}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		assert.Equal(t, updateReq.FieldsJSON, noteRes.FieldsJSON)

		// Add Tag
		addTagReq := request.AddTagRequest{Tag: "new-tag"}
		b, _ = json.Marshal(addTagReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteID, 10)+"/tags", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Remove Tag
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/notes/"+strconv.FormatInt(noteID, 10)+"/tags/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// List Notes
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var noteListRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteListRes)
		assert.NotEmpty(t, noteListRes)

		// List Notes with Filter by Deck
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?deck_id="+strconv.FormatInt(defaultDeckID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var deckFilteredRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &deckFilteredRes)
		assert.NotEmpty(t, deckFilteredRes)

		// List Notes with Filter by Tag
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=integration", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var tagFilteredRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &tagFilteredRes)
		assert.NotEmpty(t, tagFilteredRes)

		// Find Note by ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		assert.Equal(t, noteID, noteRes.ID)

		// Verify that cards were created in the database
		var cardCount int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE note_id = $1", noteID).Scan(&cardCount)
		require.NoError(t, err)
		assert.Greater(t, cardCount, 0, "Cards should have been created for the note")

		// Delete Note
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Cross-User Isolation: User B tries to create a note in User A's deck
		loginResB := registerAndLogin(t, e, "userB@example.com", "password123")
		tokenB := loginResB.AccessToken

		// Create Note in User A's deck (defaultDeckID)
		badCreateNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID, // NoteType created by User A
			DeckID:     defaultDeckID, // Deck created by User A
			FieldsJSON: `{"Front": "Bad", "Back": "Bad"}`,
		}
		b, _ = json.Marshal(badCreateNoteReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Should return 404 because deck (or note type) is not found for User B
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
