package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
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
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/dicontainer"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	"github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

func TestOwnership_Validation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg, _ := config.Load()
	cfg.Metrics.Enabled = false
	cfg.Tracing.Enabled = false
	
	logger := slog.Default()
	rdb, err := redis.NewRedisRepository(cfg.Redis, logger)
	require.NoError(t, err, "Failed to create Redis connection")
	defer rdb.Close()

	// Initialize Event Bus
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, logger)
	err = eventBus.Start()
	require.NoError(t, err, "Failed to start event bus")
	defer eventBus.Stop()

	jwtSvc, _ := jwt.NewJWTService(cfg.JWT)
	dicontainer.Init(db, rdb, eventBus, jwtSvc, cfg, logger)

	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, rdb)
	router.Init()

	// 1. Create User A and User B
	userA := registerAndLogin(t, e, "ownerA@example.com", "Password123!")
	userB := registerAndLogin(t, e, "ownerB@example.com", "Password123!")

	// --- Deck Isolation ---
	t.Run("Deck Isolation", func(t *testing.T) {
		// User A creates a deck
		deckAReq := request.CreateDeckRequest{Name: "User A Deck"}
		b, _ := json.Marshal(deckAReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		var deckA response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &deckA)

		// User B tries to GET User A's deck
		req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+strconv.FormatInt(deckA.ID, 10), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to find User A's deck")

		// User B tries to UPDATE User A's deck
		updateReq := request.UpdateDeckRequest{Name: "Hacked Deck"}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/decks/"+strconv.FormatInt(deckA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to update User A's deck")

		// User B tries to DELETE User A's deck
		deleteReq := request.DeleteDeckRequest{Action: request.ActionDeleteCards}
		b, _ = json.Marshal(deleteReq)
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/decks/"+strconv.FormatInt(deckA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to delete User A's deck")
	})

	// --- Note & Card Isolation ---
	t.Run("Note and Card Isolation", func(t *testing.T) {
		// User A setup: NoteType, Deck, Note
		// Create NoteType
		ntReq := request.CreateNoteTypeRequest{
			Name:           "User A NoteType",
			FieldsJSON:     `["Front", "Back"]`,
			CardTypesJSON:  `[{"Name": "Card 1"}]`,
			TemplatesJSON:  `[{"Front": "{{Front}}", "Back": "{{Back}}"}]`,
		}
		b, _ := json.Marshal(ntReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/note-types", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var ntA response.NoteTypeResponse
		json.Unmarshal(rec.Body.Bytes(), &ntA)

		// Create Deck
		deckAReq := request.CreateDeckRequest{Name: "User A Note Deck"}
		b, _ = json.Marshal(deckAReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var deckA response.DeckResponse
		json.Unmarshal(rec.Body.Bytes(), &deckA)

		// Create Note (this also creates cards)
		noteAReq := request.CreateNoteRequest{
			NoteTypeID: ntA.ID,
			DeckID:     deckA.ID,
			FieldsJSON: `{"Front": "A", "Back": "B"}`,
			Tags:       []string{"tagA"},
		}
		b, _ = json.Marshal(noteAReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var noteA response.NoteResponse
		json.Unmarshal(rec.Body.Bytes(), &noteA)

		// Get User A's cards to find one
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/decks/%d/cards", deckA.ID), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var cardsA []response.CardResponse
		json.Unmarshal(rec.Body.Bytes(), &cardsA)
		require.NotEmpty(t, cardsA)
		cardA := cardsA[0]

		// User B tries to GET User A's note
		req = httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+strconv.FormatInt(noteA.ID, 10), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to find User A's note")

		// User B tries to GET User A's card
		req = httptest.NewRequest(http.MethodGet, "/api/v1/cards/"+strconv.FormatInt(cardA.ID, 10), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to find User A's card")

		// User B tries to Suspend User A's card
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/cards/%d/suspend", cardA.ID), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to suspend User A's card")
	})

	// --- Media Isolation ---
	t.Run("Media Isolation", func(t *testing.T) {
		// User A creates media
		mediaReq := request.CreateMediaRequest{
			Filename:    "test.png",
			Hash:        "hashA",
			Size:        100,
			MimeType:    "image/png",
			StoragePath: "/path/a",
		}
		b, _ := json.Marshal(mediaReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/media", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var mediaA response.MediaResponse
		json.Unmarshal(rec.Body.Bytes(), &mediaA)

		// User B tries to GET User A's media
		req = httptest.NewRequest(http.MethodGet, "/api/v1/media/"+strconv.FormatInt(mediaA.ID, 10), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to find User A's media")
	})

	// --- Add-on Isolation ---
	t.Run("Add-on Isolation", func(t *testing.T) {
		// User A installs add-on
		addonReq := request.InstallAddOnRequest{
			Code:       "addonA",
			Name:       "Addon A",
			Version:    "1.0",
			ConfigJSON: "{}",
		}
		b, _ := json.Marshal(addonReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/addons", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		// User B tries to update config of User A's add-on
		updateReq := request.UpdateAddOnConfigRequest{ConfigJSON: `{"key": "value"}`}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/addons/addonA/config", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to update User A's add-on")
	})

	// --- User Preferences Isolation ---
	t.Run("User Preferences Isolation", func(t *testing.T) {
		// User A updates preferences
		updateReq := request.UpdateUserPreferencesRequest{
			Language:        "en-US",
			Theme:           "dark",
			UISize:          1.5,
			AutoSync:        true,
			NextDayStartsAt: time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC),
		}
		b, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/user/preferences", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		// User B gets their own preferences
		req = httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var prefsB response.UserPreferencesResponse
		json.Unmarshal(rec.Body.Bytes(), &prefsB)
		assert.Equal(t, "en", prefsB.Language, "User B should have default language")
		assert.Equal(t, 1.0, prefsB.UISize, "User B should have default UI size")
	})

	// --- Shared Deck Author Isolation ---
	t.Run("Shared Deck Author Isolation", func(t *testing.T) {
		// User A creates a shared deck
		sdReq := request.CreateSharedDeckRequest{
			Name:        "Public Deck",
			PackagePath: "/path/p",
			PackageSize: 1000,
			Tags:        []string{"public"},
		}
		b, _ := json.Marshal(sdReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/decks", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userA.AccessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		var sdA response.SharedDeckResponse
		json.Unmarshal(rec.Body.Bytes(), &sdA)

		// User B tries to UPDATE User A's shared deck
		updateReq := request.UpdateSharedDeckRequest{Name: "Hacked Shared Deck"}
		b, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/api/v1/marketplace/decks/"+strconv.FormatInt(sdA.ID, 10), bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to update User A's shared deck")

		// User B tries to DELETE User A's shared deck
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/marketplace/decks/"+strconv.FormatInt(sdA.ID, 10), nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+userB.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "User B should not be able to delete User A's shared deck")
	})
}
