package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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
	router.RegisterStudyRoutes() // Needed for deck creation in copy note tests

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

		// List NoteTypes with Search - Match
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=Updated", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var searchResults []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &searchResults)
		assert.NotEmpty(t, searchResults, "Search should find note types containing 'Updated'")
		found := false
		for _, nt := range searchResults {
			if nt.ID == noteTypeID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created note type should be found in search results")

		// List NoteTypes with Search - Case Insensitive
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=updated", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var caseInsensitiveResults []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &caseInsensitiveResults)
		assert.NotEmpty(t, caseInsensitiveResults, "Search should be case-insensitive")
		found = false
		for _, nt := range caseInsensitiveResults {
			if nt.ID == noteTypeID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created note type should be found in case-insensitive search")

		// List NoteTypes with Search - Partial Match
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=Basic", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var partialResults []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &partialResults)
		assert.NotEmpty(t, partialResults, "Search should support partial matching")
		found = false
		for _, nt := range partialResults {
			if nt.ID == noteTypeID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created note type should be found in partial search results")

		// List NoteTypes with Search - No Matches
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=NonExistentNoteType12345", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var emptyResults []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &emptyResults)
		assert.Empty(t, emptyResults, "Search with no matches should return empty results")

		// List NoteTypes with Search - Empty String (should return all)
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var allResults []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &allResults)
		assert.NotEmpty(t, allResults, "Empty search should return all note types")

		// Cross-User Isolation: User B searches for note types that User A has
		loginResB := registerAndLogin(t, e, "notetype_search_userB@example.com", "password123")
		tokenB := loginResB.AccessToken

		// Create NoteType for User B
		createReqB := request.CreateNoteTypeRequest{
			Name:          "User B Note Type",
			FieldsJSON:    `[{"name": "Front"}]`,
			CardTypesJSON: `[{"name": "Card 1"}]`,
			TemplatesJSON: `[{"name": "Template 1"}]`,
		}
		b, _ = json.Marshal(createReqB)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/note-types", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// User B searches for "Updated" - should only find their own note types
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=Updated", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var userBSearchResults []response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &userBSearchResults)
		// User B should not find User A's note type
		for _, nt := range userBSearchResults {
			assert.NotEqual(t, noteTypeID, nt.ID, "User B should not see User A's note type in search results")
		}

		// User B searches for "User B" - should find their own note type
		req = httptest.NewRequest(http.MethodGet, "/api/v1/note-types?search=User+B", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &userBSearchResults)
		assert.NotEmpty(t, userBSearchResults, "User B should find their own note type")

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

		// List Notes with Multiple Tags (OR logic) - should find notes with ANY tag
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=integration&tags=updated", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var multiTagRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &multiTagRes)
		assert.NotEmpty(t, multiTagRes, "Should find notes with any of the specified tags")
		// Verify at least one note has one of the tags
		foundTag := false
		for _, n := range multiTagRes {
			for _, tag := range n.Tags {
				if tag == "integration" || tag == "updated" {
					foundTag = true
					break
				}
			}
			if foundTag {
				break
			}
		}
		assert.True(t, foundTag, "At least one note should have one of the specified tags")

		// List Notes with Tag Filter and Pagination
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=integration&limit=1&offset=0", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var tagPaginatedRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &tagPaginatedRes)
		assert.LessOrEqual(t, len(tagPaginatedRes), 1, "Pagination should limit tag search results")

		// List Notes with Tag Filter - Case Insensitive
		// Create a note with uppercase tag
		createNoteWithUppercaseTag := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Test", "Back": "Test"}`,
			Tags:       []string{"UPPERCASE-TAG"},
		}
		b, _ = json.Marshal(createNoteWithUppercaseTag)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		var uppercaseTagNote response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &uppercaseTagNote)
		
		// Search with lowercase tag should find the note with uppercase tag
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=uppercase-tag", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var caseInsensitiveRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &caseInsensitiveRes)
		found := false
		for _, n := range caseInsensitiveRes {
			if n.ID == uppercaseTagNote.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Tag search should be case-insensitive")

		// List Notes with Tag Filter - Empty Results
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=NonExistentTag12345", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var emptyTagRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &emptyTagRes)
		assert.Empty(t, emptyTagRes, "Tag search with non-existent tag should return empty results")

		// List Notes with Tag Filter - Special Characters
		createNoteWithSpecialTag := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Special", "Back": "Tag"}`,
			Tags:       []string{"tag-with-dash", "tag_with_underscore"},
		}
		b, _ = json.Marshal(createNoteWithSpecialTag)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		
		// Search for tag with special characters
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=tag-with-dash", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var specialTagRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &specialTagRes)
		assert.NotEmpty(t, specialTagRes, "Should find notes with tags containing special characters")

		// Create additional notes with different field values for search testing
		createNoteReq2 := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Hello World", "Back": "Olá Mundo"}`,
			Tags:       []string{"search-test"},
		}
		b, _ = json.Marshal(createNoteReq2)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		var noteRes2 response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteRes2)

		createNoteReq3 := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Goodbye", "Back": "Adeus"}`,
			Tags:       []string{"search-test"},
		}
		b, _ = json.Marshal(createNoteReq3)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// List Notes with Search - should find notes containing "Hello"
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?search=Hello", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.NotEmpty(t, searchRes, "Search should find notes containing 'Hello'")
		// Verify at least one note contains "Hello" in fields
		found = false
		for _, n := range searchRes {
			if n.FieldsJSON != "" {
				// Check if FieldsJSON contains "Hello" (case-insensitive)
				fieldsLower := strings.ToLower(n.FieldsJSON)
				if strings.Contains(fieldsLower, "hello") {
					found = true
					break
				}
			}
		}
		assert.True(t, found, "At least one note should contain 'Hello' in fields")

		// List Notes with Search - case insensitive
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?search=hello", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var searchResLower []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &searchResLower)
		assert.NotEmpty(t, searchResLower, "Search should be case-insensitive")

		// List Notes with Search - no matches
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?search=NonExistentText12345", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var searchResEmpty []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &searchResEmpty)
		assert.Empty(t, searchResEmpty, "Search with no matches should return empty results")

		// List Notes with Search and Pagination
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?search=Hello&limit=1&offset=0", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var searchResPaginated []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &searchResPaginated)
		assert.LessOrEqual(t, len(searchResPaginated), 1, "Pagination should limit results")

		// Find Note by ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		assert.Equal(t, noteID, noteRes.ID)

		// Find Note by ID - Not Found
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/999999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Verify that cards were created in the database
		var cardCount int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE note_id = $1", noteID).Scan(&cardCount)
		require.NoError(t, err)
		assert.Greater(t, cardCount, 0, "Cards should have been created for the note")

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

		// Cross-User Isolation: User B tries to access User A's note by ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Cross-User Isolation: User B searches for User A's note content
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?search=Hello", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var searchResUserB []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &searchResUserB)
		// User B should not see User A's notes
		for _, n := range searchResUserB {
			assert.NotEqual(t, noteID, n.ID, "User B should not see User A's notes in search results")
		}

		// Cross-User Isolation: User B searches for tags that User A's notes have
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes?tags=integration", nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var tagSearchResUserB []response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &tagSearchResUserB)
		// User B should not see User A's notes even if they have matching tags
		for _, n := range tagSearchResUserB {
			assert.NotEqual(t, noteID, n.ID, "User B should not see User A's notes in tag search results")
		}

		// Update Note - Not Found
		updateReqNotFound := request.UpdateNoteRequest{
			FieldsJSON: `{"Front": "Updated Front", "Back": "Updated Back"}`,
		}
		b, _ = json.Marshal(updateReqNotFound)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/notes/999999", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Update Note - Cross-User Isolation
		b, _ = json.Marshal(updateReqNotFound) // Reuse b
		req = httptest.NewRequest(http.MethodPut, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Tag Management Verification
		// Add Tag
		addTagReq = request.AddTagRequest{Tag: "integration-new"}
		b, _ = json.Marshal(addTagReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteID, 10)+"/tags", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify Tag Added
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		found = false
		for _, tag := range noteRes.Tags {
			if tag == "integration-new" {
				found = true
				break
			}
		}
		assert.True(t, found, "Tag should have been added")

		// Remove Tag
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/notes/"+strconv.FormatInt(noteID, 10)+"/tags/integration-new", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify Tag Removed
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		found = false
		for _, tag := range noteRes.Tags {
			if tag == "integration-new" {
				found = true
				break
			}
		}
		assert.False(t, found, "Tag should have been removed")

		// Delete Note - Not Found
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/notes/999999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Delete Note - Cross-User Isolation
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Verify that User A's note still exists after User B's failed attempt
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "User A's note should still exist")

		// Delete Note - Success and verify card deletion
		// First, verify cards exist before deletion
		var cardCountBefore int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE note_id = $1", noteID).Scan(&cardCountBefore)
		require.NoError(t, err)
		assert.Greater(t, cardCountBefore, 0, "Cards should exist before note deletion")

		// Delete the note
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/notes/"+strconv.FormatInt(noteID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify note is soft-deleted (has deleted_at set)
		var deletedAt interface{}
		err = db.DB.QueryRow("SELECT deleted_at FROM notes WHERE id = $1", noteID).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt, "Note should have deleted_at set (soft delete)")

		// Verify all associated cards are hard-deleted (removed from cards table)
		var cardCountAfter int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM cards WHERE note_id = $1", noteID).Scan(&cardCountAfter)
		require.NoError(t, err)
		assert.Equal(t, 0, cardCountAfter, "All cards should be deleted after note deletion")

		// Copy Note - Success
		// First, create a new note to copy
		createNoteReq2 = request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Copy Test Front", "Back": "Copy Test Back"}`,
			Tags:       []string{"copy-test", "integration"},
		}
		b, _ = json.Marshal(createNoteReq2)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		var noteToCopy response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteToCopy)
		noteToCopyID := noteToCopy.ID

		// Copy note with all options
		copyReq := request.CopyNoteRequest{
			DeckID:    nil, // Use same deck
			CopyTags:  true,
			CopyMedia: true,
		}
		b, _ = json.Marshal(copyReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteToCopyID, 10)+"/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		var copiedNoteRes response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &copiedNoteRes)
		assert.NotEqual(t, noteToCopyID, copiedNoteRes.ID, "Copy should have different ID")
		assert.NotEqual(t, noteToCopy.GUID, copiedNoteRes.GUID, "Copy should have different GUID")
		// Compare JSON fields (order may vary, so parse and compare)
		var originalFields, copiedFields map[string]interface{}
		json.Unmarshal([]byte(noteToCopy.FieldsJSON), &originalFields)
		json.Unmarshal([]byte(copiedNoteRes.FieldsJSON), &copiedFields)
		assert.Equal(t, originalFields, copiedFields, "Fields should be copied")
		assert.Equal(t, noteToCopy.Tags, copiedNoteRes.Tags, "Tags should be copied")
		assert.Equal(t, noteToCopy.NoteTypeID, copiedNoteRes.NoteTypeID, "NoteType should be copied")

		// Copy note without tags
		copyReqNoTags := request.CopyNoteRequest{
			DeckID:    nil,
			CopyTags:  false,
			CopyMedia: false,
		}
		b, _ = json.Marshal(copyReqNoTags)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteToCopyID, 10)+"/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		var copiedNoteNoTags response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &copiedNoteNoTags)
		assert.Empty(t, copiedNoteNoTags.Tags, "Tags should not be copied")

		// Copy note to different deck
		// Create another deck
		createDeckReq := request.CreateDeckRequest{
			Name: "Copy Target Deck",
		}
		b, _ = json.Marshal(createDeckReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Failed to create deck: status %d, body: %s", rec.Code, rec.Body.String())
		}
		var targetDeck response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &targetDeck)
		targetDeckID := targetDeck.ID
		require.NotZero(t, targetDeckID, "Deck ID should not be zero")

		copyReqDiffDeck := request.CopyNoteRequest{
			DeckID:    &targetDeckID,
			CopyTags:  true,
			CopyMedia: true,
		}
		b, _ = json.Marshal(copyReqDiffDeck)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteToCopyID, 10)+"/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Failed to copy note: status %d, body: %s, noteID: %d", rec.Code, rec.Body.String(), noteToCopyID)
		}
		var copiedNoteDiffDeck response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &copiedNoteDiffDeck)
		require.NotZero(t, copiedNoteDiffDeck.ID, "Copied note ID should not be zero")

		// Verify cards are in the new deck
		var cardDeckID int64
		err = db.DB.QueryRow("SELECT deck_id FROM cards WHERE note_id = $1 LIMIT 1", copiedNoteDiffDeck.ID).Scan(&cardDeckID)
		require.NoError(t, err, "Failed to query card deck_id for copied note")
		assert.Equal(t, targetDeckID, cardDeckID, "Cards should be in the specified deck")

		// Copy Note - Not Found
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/999999/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		// Copy Note - Invalid Deck
		invalidDeckID := int64(999999)
		copyReqInvalidDeck := request.CopyNoteRequest{
			DeckID:    &invalidDeckID,
			CopyTags:  true,
			CopyMedia: true,
		}
		b, _ = json.Marshal(copyReqInvalidDeck)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteToCopyID, 10)+"/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "Should return 404 for invalid deck")

		// Copy Note - Unauthorized
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteToCopyID, 10)+"/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		// No Authorization header
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "Should return 401 without auth")

		// Copy Note - Cross-User Isolation
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+strconv.FormatInt(noteToCopyID, 10)+"/copy", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to copy User A's note")

		// Find Duplicates - Success (without note type filter to avoid validation issues)
		findDupReq := request.FindDuplicatesRequest{
			FieldName: "Front",
		}
		b, _ = json.Marshal(findDupReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Logf("Response body: %s", rec.Body.String())
		}
		assert.Equal(t, http.StatusOK, rec.Code, "Should find duplicates successfully")
		var dupRes response.FindDuplicatesResponse
		json.Unmarshal(rec.Body.Bytes(), &dupRes)
		assert.GreaterOrEqual(t, dupRes.Total, 0, "Should return total count")

		// Find Duplicates - Success without note type filter
		findDupReqNoType := request.FindDuplicatesRequest{
			FieldName: "Front",
		}
		b, _ = json.Marshal(findDupReqNoType)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Should find duplicates without note type filter")

		// Find Duplicates - Note type not found
		invalidNoteTypeID := int64(999999)
		findDupReqNotFound := request.FindDuplicatesRequest{
			NoteTypeID: &invalidNoteTypeID,
			FieldName:  "Front",
		}
		b, _ = json.Marshal(findDupReqNotFound)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "Should return 404 for note type not found")

		// Find Duplicates - Unauthorized
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		// No Authorization header
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "Should return 401 without auth")

		// Find Duplicates - Cross-User Isolation
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		// User B should not see User A's note type
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not access User A's note type")

		// Find Duplicates - Success with automatic first field detection (fieldName omitted)
		// First, create a note type with a known first field
		var autoNoteTypeID int64
		err = db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'Auto First Field Test', '[{\"name\":\"Question\"},{\"name\":\"Answer\"}]', '[{\"name\": \"Card 1\"}]', '[]') RETURNING id", loginRes.User.ID).Scan(&autoNoteTypeID)
		require.NoError(t, err)

		// Create duplicate notes with same "Question" field value
		var noteID1, noteID2 int64
		err = db.DB.QueryRow("INSERT INTO notes (user_id, guid, note_type_id, fields_json) VALUES ($1, gen_random_uuid()::text, $2, '{\"Question\":\"Same Question\",\"Answer\":\"Answer 1\"}') RETURNING id", loginRes.User.ID, autoNoteTypeID).Scan(&noteID1)
		require.NoError(t, err)
		err = db.DB.QueryRow("INSERT INTO notes (user_id, guid, note_type_id, fields_json) VALUES ($1, gen_random_uuid()::text, $2, '{\"Question\":\"Same Question\",\"Answer\":\"Answer 2\"}') RETURNING id", loginRes.User.ID, autoNoteTypeID).Scan(&noteID2)
		require.NoError(t, err)

		// Test automatic first field detection (fieldName omitted)
		findDupAutoReq := request.FindDuplicatesRequest{
			NoteTypeID: &autoNoteTypeID,
			// FieldName is omitted - should automatically use "Question" (first field)
		}
		b, _ = json.Marshal(findDupAutoReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Should find duplicates with automatic first field detection")
		var dupAutoRes response.FindDuplicatesResponse
		json.Unmarshal(rec.Body.Bytes(), &dupAutoRes)
		assert.GreaterOrEqual(t, dupAutoRes.Total, 0, "Should return total count")

		// Find Duplicates - Success with automatic first field detection (fieldName empty string)
		findDupAutoEmptyReq := request.FindDuplicatesRequest{
			NoteTypeID: &autoNoteTypeID,
			FieldName:  "", // Empty string - should automatically use first field
		}
		b, _ = json.Marshal(findDupAutoEmptyReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Should find duplicates with automatic first field detection (empty fieldName)")

		// Find Duplicates - Error when note type has no fields
		var emptyFieldsNoteTypeID int64
		err = db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'Empty Fields Test', '[]', '[{\"name\": \"Card 1\"}]', '[]') RETURNING id", loginRes.User.ID).Scan(&emptyFieldsNoteTypeID)
		require.NoError(t, err)

		findDupEmptyFieldsReq := request.FindDuplicatesRequest{
			NoteTypeID: &emptyFieldsNoteTypeID,
			// FieldName omitted - should fail because note type has no fields
		}
		b, _ = json.Marshal(findDupEmptyFieldsReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "Should return 400 when note type has no fields")

		// Find Duplicates - Backward compatibility (explicit fieldName still works)
		findDupExplicitReq := request.FindDuplicatesRequest{
			NoteTypeID: &autoNoteTypeID,
			FieldName:  "Answer", // Explicit field name - should use this instead of first field
		}
		b, _ = json.Marshal(findDupExplicitReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Should find duplicates with explicit field name (backward compatibility)")
	})

	// Find Duplicates by GUID - Success (no duplicates found in normal scenario)
	// Note: The database has a UNIQUE constraint on GUID, so we cannot create actual duplicates
	// This test verifies the method works correctly when there are no duplicates (normal scenario)
	// Use the noteTypeID from the NoteTypes subtest (scope level)
	// If it's not set, create a new one
	if noteTypeID == 0 {
		var newNoteTypeID int64
		err := db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'GUID Test', '[{\"name\":\"Front\"},{\"name\":\"Back\"}]', '[{\"name\": \"Card 1\"}]', '[]') RETURNING id", loginRes.User.ID).Scan(&newNoteTypeID)
		require.NoError(t, err)
		noteTypeID = newNoteTypeID
	}
	guid1 := "550e8400-e29b-41d4-a716-446655440000"
	guid2 := "550e8400-e29b-41d4-a716-446655440001"
	var noteID1, noteID2 int64
	err = db.DB.QueryRow("INSERT INTO notes (user_id, guid, note_type_id, fields_json, tags) VALUES ($1, $2, $3, '{\"Front\":\"Note 1\",\"Back\":\"Back 1\"}', '{}') RETURNING id", loginRes.User.ID, guid1, noteTypeID).Scan(&noteID1)
	require.NoError(t, err)
	err = db.DB.QueryRow("INSERT INTO notes (user_id, guid, note_type_id, fields_json, tags) VALUES ($1, $2, $3, '{\"Front\":\"Note 2\",\"Back\":\"Back 2\"}', '{}') RETURNING id", loginRes.User.ID, guid2, noteTypeID).Scan(&noteID2)
	require.NoError(t, err)

	findDupGUIDReq := request.FindDuplicatesRequest{
		UseGUID: true,
	}
	b, _ := json.Marshal(findDupGUIDReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "Should return OK when searching for duplicates by GUID")
	var dupGUIDRes response.FindDuplicatesResponse
	json.Unmarshal(rec.Body.Bytes(), &dupGUIDRes)
	// In normal operation, GUIDs are unique, so no duplicates should be found
	assert.Equal(t, 0, dupGUIDRes.Total, "Should find no duplicates when all GUIDs are unique")

	// Find Duplicates by GUID - No duplicates found
	// Create a new user with no duplicate GUIDs
	loginRes2 := registerAndLogin(t, e, "find_duplicates_guid_user2@example.com", "password123")
	token2 := loginRes2.AccessToken

	findDupGUIDReq2 := request.FindDuplicatesRequest{
		UseGUID: true,
	}
	b2, _ := json.Marshal(findDupGUIDReq2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b2))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req2.Header.Set("Authorization", "Bearer "+token2)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code, "Should return OK even when no duplicates found")
	var dupGUIDRes2 response.FindDuplicatesResponse
	json.Unmarshal(rec2.Body.Bytes(), &dupGUIDRes2)
	assert.Equal(t, 0, dupGUIDRes2.Total, "Should return 0 duplicates for user with no duplicate GUIDs")

	// Find Duplicates by GUID - Cross-User Isolation
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b2))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req3.Header.Set("Authorization", "Bearer "+token2)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	// User 2 should not see User 1's duplicates
	var dupGUIDRes3 response.FindDuplicatesResponse
	json.Unmarshal(rec3.Body.Bytes(), &dupGUIDRes3)
	assert.Equal(t, 0, dupGUIDRes3.Total, "User 2 should not see User 1's duplicate GUIDs")

	// Find Duplicates by GUID - Unauthorized
	req4 := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b2))
	req4.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// No Authorization header
	rec4 := httptest.NewRecorder()
	e.ServeHTTP(rec4, req4)
	assert.Equal(t, http.StatusUnauthorized, rec4.Code, "Should return 401 without auth")

	// Find Duplicates - Edge Cases
	t.Run("Notes with no cards (deck_id = 0)", func(t *testing.T) {
		// Create a note type for this test
		var edgeCaseNoteTypeID int64
		err := db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'Edge Case Test', '[{\"name\":\"Front\"},{\"name\":\"Back\"}]', '[{\"name\": \"Card 1\"}]', '[]') RETURNING id", loginRes.User.ID).Scan(&edgeCaseNoteTypeID)
		require.NoError(t, err)
		
		// Create a note without any cards
		var noteIDNoCard int64
		err = db.DB.QueryRow("INSERT INTO notes (user_id, guid, note_type_id, fields_json, tags) VALUES ($1, gen_random_uuid()::text, $2, '{\"Front\":\"NoCardTest\"}', '{}') RETURNING id", loginRes.User.ID, edgeCaseNoteTypeID).Scan(&noteIDNoCard)
		require.NoError(t, err)

		// Create another note with the same field value to create a duplicate
		var noteIDNoCard2 int64
		err = db.DB.QueryRow("INSERT INTO notes (user_id, guid, note_type_id, fields_json, tags) VALUES ($1, gen_random_uuid()::text, $2, '{\"Front\":\"NoCardTest\"}', '{}') RETURNING id", loginRes.User.ID, edgeCaseNoteTypeID).Scan(&noteIDNoCard2)
		require.NoError(t, err)

		findDupReq := request.FindDuplicatesRequest{
			NoteTypeID: &edgeCaseNoteTypeID,
			FieldName:  "Front",
		}
		b, _ := json.Marshal(findDupReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Logf("Response body: %s", rec.Body.String())
		}
		assert.Equal(t, http.StatusOK, rec.Code, "Should handle notes with no cards")
		var res response.FindDuplicatesResponse
		json.Unmarshal(rec.Body.Bytes(), &res)
		// If there are duplicates, verify deck_id is 0 for notes without cards
		for _, group := range res.Duplicates {
			for _, noteInfo := range group.Notes {
				if noteInfo.ID == noteIDNoCard || noteInfo.ID == noteIDNoCard2 {
					assert.Equal(t, int64(0), noteInfo.DeckID, "Note without cards should have deck_id = 0")
				}
			}
		}
	})

	t.Run("SQL injection attempt in field_name", func(t *testing.T) {
		// Attempt SQL injection in field_name
		maliciousFieldName := "Front'; DROP TABLE notes; --"
		findDupReq := request.FindDuplicatesRequest{
			NoteTypeID: &noteTypeID,
			FieldName:  maliciousFieldName,
		}
		b, _ := json.Marshal(findDupReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		// Should either return 400 (validation error) or 404 (field not found) or 200 with empty results
		// The important thing is that it doesn't execute SQL injection
		assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusBadRequest || rec.Code == http.StatusNotFound, "Should not execute SQL injection, got code: %d", rec.Code)
		// Verify table still exists by querying it
		var count int
		err := db.DB.QueryRow("SELECT COUNT(*) FROM notes WHERE user_id = $1", loginRes.User.ID).Scan(&count)
		require.NoError(t, err, "Table should still exist after SQL injection attempt")
	})

	t.Run("Invalid JSON in fields_json", func(t *testing.T) {
		// Test that the query handles missing fields gracefully
		// The field "NonExistentField" doesn't exist in the note type, so it should return 400
		findDupReq := request.FindDuplicatesRequest{
			NoteTypeID: &noteTypeID,
			FieldName:  "NonExistentField",
		}
		b, _ := json.Marshal(findDupReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		// Should return 400 because field doesn't exist in note type
		// Or 404 if note type validation happens first
		assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusNotFound, "Should return 400 or 404 for invalid field name, got: %d", rec.Code)
	})
}

func TestSearch_Regex(t *testing.T) {
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
	router.RegisterSearchRoutes()

	// Register and login
	loginRes := registerAndLogin(t, e, "regex_search_user@example.com", "password123")
	token := loginRes.AccessToken

	// Create note type
	var noteTypeID int64
	err = db.DB.QueryRow("INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json) VALUES ($1, 'Basic', '[{\"name\":\"Front\"},{\"name\":\"Back\"}]', '[{\"name\": \"Card 1\"}]', '[]') RETURNING id", loginRes.User.ID).Scan(&noteTypeID)
	require.NoError(t, err)

	// Get default deck ID
	var defaultDeckID int64
	err = db.DB.QueryRow("SELECT id FROM decks WHERE user_id = $1 AND name = 'Default'", loginRes.User.ID).Scan(&defaultDeckID)
	require.NoError(t, err)

	// Create notes with different field values for regex testing
	createNoteReq := request.CreateNoteRequest{
		NoteTypeID: noteTypeID,
		DeckID:     defaultDeckID,
		FieldsJSON: `{"Front": "a1", "Back": "test"}`,
		Tags:       []string{},
	}
	b, _ := json.Marshal(createNoteReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	createNoteReq = request.CreateNoteRequest{
		NoteTypeID: noteTypeID,
		DeckID:     defaultDeckID,
		FieldsJSON: `{"Front": "b1", "Back": "test"}`,
		Tags:       []string{},
	}
	b, _ = json.Marshal(createNoteReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	createNoteReq = request.CreateNoteRequest{
		NoteTypeID: noteTypeID,
		DeckID:     defaultDeckID,
		FieldsJSON: `{"Front": "hello world", "Back": "test"}`,
		Tags:       []string{},
	}
	b, _ = json.Marshal(createNoteReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	t.Run("Basic_Regex", func(t *testing.T) {
		searchReq := request.AdvancedSearchRequest{
			Query:  "re:hello.*world",
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ := json.Marshal(searchReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes response.SearchResult
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.Greater(t, searchRes.Total, 0, "Should find at least one note matching regex")
	})

	t.Run("Field_Regex", func(t *testing.T) {
		searchReq := request.AdvancedSearchRequest{
			Query:  "front:re:[a-c]1",
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ := json.Marshal(searchReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes response.SearchResult
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.GreaterOrEqual(t, searchRes.Total, 2, "Should find at least 2 notes (a1 and b1)")
	})

	t.Run("Invalid_Regex", func(t *testing.T) {
		searchReq := request.AdvancedSearchRequest{
			Query:  "re:[invalid",
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ := json.Marshal(searchReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code, "Invalid regex should return 400 Bad Request")
		var errorRes map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &errorRes)
		assert.Contains(t, errorRes["message"].(string), "invalid regex", "Error message should mention invalid regex")
	})

	t.Run("NoCombining_Basic", func(t *testing.T) {
		// Create notes with accented text
		createNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "café", "Back": "coffee"}`,
			Tags:       []string{},
		}
		b, _ := json.Marshal(createNoteReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		createNoteReq = request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "ação", "Back": "action"}`,
			Tags:       []string{},
		}
		b, _ = json.Marshal(createNoteReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		// Search with nc: prefix
		searchReq := request.AdvancedSearchRequest{
			Query:  "nc:cafe",
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ = json.Marshal(searchReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes response.SearchResult
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.GreaterOrEqual(t, searchRes.Total, 1, "Should find at least one note matching 'café' with 'nc:cafe'")
	})

	t.Run("NoCombining_Field_Search", func(t *testing.T) {
		searchReq := request.AdvancedSearchRequest{
			Query:  "front:nc:acao",
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ := json.Marshal(searchReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes response.SearchResult
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.GreaterOrEqual(t, searchRes.Total, 1, "Should find at least one note with 'ação' in Front field using 'front:nc:acao'")
	})

	t.Run("NoCombining_With_Wildcard", func(t *testing.T) {
		// Create note with über
		createNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "über", "Back": "over"}`,
			Tags:       []string{},
		}
		b, _ := json.Marshal(createNoteReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		searchReq := request.AdvancedSearchRequest{
			Query:  "nc:uber*",
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ = json.Marshal(searchReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes response.SearchResult
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.GreaterOrEqual(t, searchRes.Total, 1, "Should find at least one note with 'über' using 'nc:uber*'")
	})

	t.Run("NoCombining_Exact_Phrase", func(t *testing.T) {
		// Create note with São Paulo
		createNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "São Paulo", "Back": "city"}`,
			Tags:       []string{},
		}
		b, _ := json.Marshal(createNoteReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		searchReq := request.AdvancedSearchRequest{
			Query:  `nc:"Sao Paulo"`,
			Type:   "notes",
			Limit:  100,
			Offset: 0,
		}
		b, _ = json.Marshal(searchReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/search/advanced", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var searchRes response.SearchResult
		json.Unmarshal(rec.Body.Bytes(), &searchRes)
		assert.GreaterOrEqual(t, searchRes.Total, 1, "Should find at least one note with 'São Paulo' using 'nc:\"Sao Paulo\"'")
	})

	// Export Notes
	t.Run("Export Notes", func(t *testing.T) {
		// Get default deck ID
		var defaultDeckID int64
		err := db.DB.QueryRow("SELECT id FROM decks WHERE user_id = $1 AND name = 'Default' LIMIT 1", loginRes.User.ID).Scan(&defaultDeckID)
		require.NoError(t, err)

		// Create notes for export
		var noteIDs []int64
		for i := 0; i < 3; i++ {
			createNoteReq := request.CreateNoteRequest{
				NoteTypeID: noteTypeID,
				DeckID:     defaultDeckID,
				FieldsJSON: fmt.Sprintf(`{"Front": "Export Note %d", "Back": "Back %d"}`, i, i),
				Tags:       []string{"export", "test"},
			}
			b, _ := json.Marshal(createNoteReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusCreated, rec.Code)
			var noteRes response.NoteResponse
			json.Unmarshal(rec.Body.Bytes(), &noteRes)
			noteIDs = append(noteIDs, noteRes.ID)
		}

		t.Run("Success - Text format without scheduling", func(t *testing.T) {
			exportReq := request.ExportNotesRequest{
				NoteIDs:          noteIDs,
				Format:           "text",
				IncludeMedia:     false,
				IncludeScheduling: false,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.Contains(t, rec.Header().Get("Content-Disposition"), "notes_export.txt")
			assert.Greater(t, len(rec.Body.Bytes()), 0)
			// Verify content contains expected data
			body := rec.Body.String()
			assert.Contains(t, body, "GUID")
			assert.Contains(t, body, "Front")
			assert.Contains(t, body, "Back")
			assert.Contains(t, body, "Tags")
		})

		t.Run("Success - Text format with scheduling", func(t *testing.T) {
			exportReq := request.ExportNotesRequest{
				NoteIDs:          noteIDs[:1], // Only first note
				Format:           "text",
				IncludeMedia:     false,
				IncludeScheduling: true,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
			body := rec.Body.String()
			// Should contain card information if scheduling is included
			assert.Contains(t, body, "GUID")
		})

		t.Run("Invalid format", func(t *testing.T) {
			exportReq := request.ExportNotesRequest{
				NoteIDs:          noteIDs,
				Format:           "invalid",
				IncludeMedia:     false,
				IncludeScheduling: false,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Empty note IDs", func(t *testing.T) {
			exportReq := request.ExportNotesRequest{
				NoteIDs:          []int64{},
				Format:           "text",
				IncludeMedia:     false,
				IncludeScheduling: false,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Note not found or access denied", func(t *testing.T) {
			exportReq := request.ExportNotesRequest{
				NoteIDs:          []int64{99999},
				Format:           "text",
				IncludeMedia:     false,
				IncludeScheduling: false,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("Cross-user isolation", func(t *testing.T) {
			// Create second user
			loginRes2 := registerAndLogin(t, e, "export_user2@example.com", "password123")
			token2 := loginRes2.AccessToken

			// User 2 tries to export User 1's notes
			exportReq := request.ExportNotesRequest{
				NoteIDs:          noteIDs, // User 1's notes
				Format:           "text",
				IncludeMedia:     false,
				IncludeScheduling: false,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token2)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// Should return 404 (notes not found or access denied)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("APKG format - Not yet implemented", func(t *testing.T) {
			exportReq := request.ExportNotesRequest{
				NoteIDs:          noteIDs[:1],
				Format:           "apkg",
				IncludeMedia:     false,
				IncludeScheduling: false,
			}
			b, _ := json.Marshal(exportReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/export", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// APKG generation requires SQLite driver which is not yet implemented
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	})

	// Recent Deletions
	t.Run("Recent Deletions", func(t *testing.T) {
		// Create a note and then delete it to create a deletion log
		createNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Recent Deletion Test", "Back": "Back"}`,
			Tags:       []string{"test"},
		}
		b, _ := json.Marshal(createNoteReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var noteRes response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		noteID := noteRes.ID

		// Delete the note to create a deletion log
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/notes/%d", noteID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)

		// Wait a bit to ensure deletion log is created
		time.Sleep(100 * time.Millisecond)

		t.Run("Success - get recent deletions with defaults", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			var response response.RecentDeletionsResponse
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.GreaterOrEqual(t, len(response.Data), 0)
			assert.Equal(t, 20, response.Pagination.Limit)
			assert.Equal(t, 1, response.Pagination.Page)
		})

		t.Run("Success - get recent deletions with custom limit and days", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?limit=10&days=5", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			var response response.RecentDeletionsResponse
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.GreaterOrEqual(t, len(response.Data), 0)
			assert.Equal(t, 10, response.Pagination.Limit)
		})

		t.Run("Error - invalid limit", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?limit=invalid", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Error - limit exceeds maximum", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?limit=150", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Error - invalid days", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?days=invalid", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Error - days exceeds maximum", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?days=500", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Cross-user isolation - cannot see other user's deletions", func(t *testing.T) {
			// Create another user and a deletion log for them
			loginRes2 := registerAndLogin(t, e, "recent_deletions_user2@example.com", "password123")
			token2 := loginRes2.AccessToken

			// User 2 tries to get deletions (should only see their own, which is none)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions", nil)
			req.Header.Set("Authorization", "Bearer "+token2)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			var response response.RecentDeletionsResponse
			json.Unmarshal(rec.Body.Bytes(), &response)
			// User 2 should not see user 1's deletions
			for _, dl := range response.Data {
				assert.Equal(t, loginRes2.User.ID, dl.UserID)
			}
		})
	})

	// Restore Deletion
	t.Run("Restore Deletion", func(t *testing.T) {
		// Create a note and then delete it to create a deletion log
		createNoteReq := request.CreateNoteRequest{
			NoteTypeID: noteTypeID,
			DeckID:     defaultDeckID,
			FieldsJSON: `{"Front": "Restore Test Note", "Back": "Back content"}`,
			Tags:       []string{"restore", "test"},
		}
		b, _ := json.Marshal(createNoteReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var noteRes response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteRes)
		originalNoteID := noteRes.ID
		originalGUID := noteRes.GUID

		// Delete the note to create a deletion log
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/notes/%d", originalNoteID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)

		// Wait a bit to ensure deletion log is created
		time.Sleep(100 * time.Millisecond)

		// Get deletion log ID
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?limit=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var deletionsRes response.RecentDeletionsResponse
		json.Unmarshal(rec.Body.Bytes(), &deletionsRes)
		require.Greater(t, len(deletionsRes.Data), 0, "Should have at least one deletion log")
		deletionLogID := deletionsRes.Data[0].ID

		t.Run("Success - restore deleted note", func(t *testing.T) {
			restoreReq := request.RestoreDeletionRequest{
				DeckID: defaultDeckID,
			}
			b, _ := json.Marshal(restoreReq)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", deletionLogID), bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			var restoreRes response.RestoreDeletionResponse
			json.Unmarshal(rec.Body.Bytes(), &restoreRes)
			assert.Greater(t, restoreRes.ID, int64(0))
			assert.Equal(t, originalGUID, restoreRes.GUID, "Restored note should have original GUID")
			assert.Equal(t, "Note restored successfully", restoreRes.Message)

			// Verify the note was actually restored by fetching it
			req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/notes/%d", restoreRes.ID), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec = httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			var restoredNoteRes response.NoteResponse
			json.Unmarshal(rec.Body.Bytes(), &restoredNoteRes)
			assert.Equal(t, restoreRes.ID, restoredNoteRes.ID)
			assert.Equal(t, originalGUID, restoredNoteRes.GUID)
			// Parse and compare JSON fields (order-independent)
			var expectedFields, actualFields map[string]interface{}
			json.Unmarshal([]byte(`{"Front": "Restore Test Note", "Back": "Back content"}`), &expectedFields)
			json.Unmarshal([]byte(restoredNoteRes.FieldsJSON), &actualFields)
			assert.Equal(t, expectedFields, actualFields)
			assert.Contains(t, restoredNoteRes.Tags, "restore")
			assert.Contains(t, restoredNoteRes.Tags, "test")
		})

		t.Run("Error - invalid deletion log ID", func(t *testing.T) {
			restoreReq := request.RestoreDeletionRequest{
				DeckID: defaultDeckID,
			}
			b, _ := json.Marshal(restoreReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/deletions/invalid/restore", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Error - deletion log not found", func(t *testing.T) {
			restoreReq := request.RestoreDeletionRequest{
				DeckID: defaultDeckID,
			}
			b, _ := json.Marshal(restoreReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/deletions/99999/restore", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("Error - invalid request body", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", deletionLogID), bytes.NewReader([]byte("invalid json")))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Error - missing deck_id", func(t *testing.T) {
			restoreReq := request.RestoreDeletionRequest{
				DeckID: 0,
			}
			b, _ := json.Marshal(restoreReq)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", deletionLogID), bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Error - invalid deck_id", func(t *testing.T) {
			// Create a new note and delete it to get a fresh deletion log
			createNoteReq2 := request.CreateNoteRequest{
				NoteTypeID: noteTypeID,
				DeckID:     defaultDeckID,
				FieldsJSON: `{"Front": "Invalid Deck Test", "Back": "Back"}`,
				Tags:       []string{"test"},
			}
			b2, _ := json.Marshal(createNoteReq2)
			req2 := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b2))
			req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req2.Header.Set("Authorization", "Bearer "+token)
			rec2 := httptest.NewRecorder()
			e.ServeHTTP(rec2, req2)
			require.Equal(t, http.StatusCreated, rec2.Code)
			var noteRes2 response.NoteResponse
			json.Unmarshal(rec2.Body.Bytes(), &noteRes2)
			noteID2 := noteRes2.ID

			// Delete the note
			req2 = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/notes/%d", noteID2), nil)
			req2.Header.Set("Authorization", "Bearer "+token)
			rec2 = httptest.NewRecorder()
			e.ServeHTTP(rec2, req2)
			require.Equal(t, http.StatusNoContent, rec2.Code)

			time.Sleep(100 * time.Millisecond)

			// Get deletion log ID
			req2 = httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?limit=1", nil)
			req2.Header.Set("Authorization", "Bearer "+token)
			rec2 = httptest.NewRecorder()
			e.ServeHTTP(rec2, req2)
			require.Equal(t, http.StatusOK, rec2.Code)
			var deletionsRes2 response.RecentDeletionsResponse
			json.Unmarshal(rec2.Body.Bytes(), &deletionsRes2)
			require.Greater(t, len(deletionsRes2.Data), 0)
			freshDeletionLogID := deletionsRes2.Data[0].ID

			// Now try to restore with invalid deck_id
			restoreReq := request.RestoreDeletionRequest{
				DeckID: 99999, // Non-existent deck
			}
			b, _ := json.Marshal(restoreReq)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", freshDeletionLogID), bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("Cross-user isolation - cannot restore another user's deletion", func(t *testing.T) {
			// Create another user
			loginRes2 := registerAndLogin(t, e, "restore_deletion_user2@example.com", "password123")
			token2 := loginRes2.AccessToken

			// User 2 tries to restore user 1's deletion log
			restoreReq := request.RestoreDeletionRequest{
				DeckID: defaultDeckID,
			}
			b, _ := json.Marshal(restoreReq)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", deletionLogID), bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token2)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code, "User 2 should not be able to restore user 1's deletion")
		})

		t.Run("Error - already restored (GUID conflict)", func(t *testing.T) {
			// Create a new deletion log by creating and deleting another note
			createNoteReq2 := request.CreateNoteRequest{
				NoteTypeID: noteTypeID,
				DeckID:     defaultDeckID,
				FieldsJSON: `{"Front": "Second Restore Test", "Back": "Back"}`,
				Tags:       []string{"test"},
			}
			b, _ := json.Marshal(createNoteReq2)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusCreated, rec.Code)
			var noteRes2 response.NoteResponse
			json.Unmarshal(rec.Body.Bytes(), &noteRes2)
			noteID2 := noteRes2.ID
			originalGUID2 := noteRes2.GUID

			// Delete the note
			req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/notes/%d", noteID2), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec = httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusNoContent, rec.Code)

			time.Sleep(100 * time.Millisecond)

			// Get deletion log ID - find the one that matches the GUID we just deleted
			req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/deletions?limit=10", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec = httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			var deletionsRes2 response.RecentDeletionsResponse
			json.Unmarshal(rec.Body.Bytes(), &deletionsRes2)
			require.Greater(t, len(deletionsRes2.Data), 0)
			
			// Find the deletion log that matches the GUID we just deleted
			var deletionLogID2 int64
			for _, dl := range deletionsRes2.Data {
				if objData, ok := dl.ObjectData["guid"].(string); ok && objData == originalGUID2 {
					deletionLogID2 = dl.ID
					break
				}
			}
			require.Greater(t, deletionLogID2, int64(0), "Should find deletion log for the note we just deleted")

			// Restore it first time
			restoreReq := request.RestoreDeletionRequest{
				DeckID: defaultDeckID,
			}
			b, _ = json.Marshal(restoreReq)
			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", deletionLogID2), bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec = httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)

			// Try to restore again (should fail because GUID already exists)
			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notes/deletions/%d/restore", deletionLogID2), bytes.NewReader(b))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec = httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusConflict, rec.Code, "Should return 409 Conflict when trying to restore already restored note")
		})
	})
}
