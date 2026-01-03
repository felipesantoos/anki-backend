package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteService)
	handler := handlers.NewNoteHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateNoteRequest{
			NoteTypeID: 1,
			DeckID:     1,
			FieldsJSON: `{"Front": "cat"}`,
			Tags:       []string{"animal"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
		n, _ := note.NewBuilder().WithID(1).WithUserID(userID).WithGUID(guid).WithNoteTypeID(reqBody.NoteTypeID).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.NoteTypeID, reqBody.DeckID, reqBody.FieldsJSON, reqBody.Tags).Return(n, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestNoteHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteService)
	mockExportSvc := new(MockExportService)
	handler := handlers.NewNoteHandler(mockSvc, mockExportSvc)
	userID := int64(1)
	noteID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateNoteRequest{
			FieldsJSON: `{"Front": "dog"}`,
			Tags:       []string{"pet"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/notes/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
		n, _ := note.NewBuilder().WithID(noteID).WithUserID(userID).WithGUID(guid).WithNoteTypeID(1).Build()
		mockSvc.On("Update", mock.Anything, userID, noteID, reqBody.FieldsJSON, reqBody.Tags).Return(n, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestNoteHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteService)
	mockExportSvc := new(MockExportService)
	handler := handlers.NewNoteHandler(mockSvc, mockExportSvc)
	userID := int64(1)
	noteID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/notes/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, noteID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestNoteHandler_FindDuplicates(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteService)
	mockExportSvc := new(MockExportService)
	handler := handlers.NewNoteHandler(mockSvc, mockExportSvc)
	userID := int64(1)

	t.Run("Success with UseGUID=true", func(t *testing.T) {
		reqBody := request.FindDuplicatesRequest{
			UseGUID: true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		result := &note.DuplicateResult{
			Duplicates: []*note.DuplicateGroup{
				{
					FieldValue: "550e8400-e29b-41d4-a716-446655440000",
					Notes: []*note.DuplicateNoteInfo{
						{ID: 1, GUID: "550e8400-e29b-41d4-a716-446655440000", DeckID: 20, CreatedAt: time.Now()},
						{ID: 2, GUID: "550e8400-e29b-41d4-a716-446655440000", DeckID: 21, CreatedAt: time.Now()},
					},
				},
			},
			Total: 1,
		}

		mockSvc.On("FindDuplicatesByGUID", mock.Anything, userID).Return(result, nil).Once()

		if assert.NoError(t, handler.FindDuplicates(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.Equal(t, float64(1), response["total_duplicates"])
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with UseGUID=false and field_name", func(t *testing.T) {
		noteTypeID := int64(1)
		fieldName := "Front"
		reqBody := request.FindDuplicatesRequest{
			UseGUID:   false,
			NoteTypeID: &noteTypeID,
			FieldName: fieldName,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		result := &note.DuplicateResult{
			Duplicates: []*note.DuplicateGroup{
				{
					FieldValue: "Hello",
					Notes: []*note.DuplicateNoteInfo{
						{ID: 1, GUID: "guid1", DeckID: 20, CreatedAt: time.Now()},
						{ID: 2, GUID: "guid2", DeckID: 21, CreatedAt: time.Now()},
					},
				},
			},
			Total: 1,
		}

		mockSvc.On("FindDuplicates", mock.Anything, userID, &noteTypeID, fieldName).Return(result, nil).Once()

		if assert.NoError(t, handler.FindDuplicates(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.Equal(t, float64(1), response["total_duplicates"])
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with UseGUID=false and automatic first field", func(t *testing.T) {
		noteTypeID := int64(1)
		reqBody := request.FindDuplicatesRequest{
			UseGUID:   false,
			NoteTypeID: &noteTypeID,
			FieldName: "", // Empty field name should trigger automatic first field detection
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		result := &note.DuplicateResult{
			Duplicates: []*note.DuplicateGroup{},
			Total:      0,
		}

		// Service will automatically use first field, but we pass empty string here
		// The service will call GetFirstFieldName internally
		mockSvc.On("FindDuplicates", mock.Anything, userID, &noteTypeID, "").Return(result, nil).Once()

		if assert.NoError(t, handler.FindDuplicates(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Validation error - invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.FindDuplicates(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("Service error - note type not found", func(t *testing.T) {
		noteTypeID := int64(999)
		reqBody := request.FindDuplicatesRequest{
			UseGUID:   false,
			NoteTypeID: &noteTypeID,
			FieldName: "Front",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("FindDuplicates", mock.Anything, userID, &noteTypeID, "Front").Return(nil, fmt.Errorf("note type not found")).Once()

		err := handler.FindDuplicates(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service error - generic error", func(t *testing.T) {
		reqBody := request.FindDuplicatesRequest{
			UseGUID: true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("FindDuplicatesByGUID", mock.Anything, userID).Return(nil, fmt.Errorf("database error")).Once()

		err := handler.FindDuplicates(c)
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success - no duplicates found", func(t *testing.T) {
		reqBody := request.FindDuplicatesRequest{
			UseGUID: true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/find-duplicates", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		result := &note.DuplicateResult{
			Duplicates: []*note.DuplicateGroup{},
			Total:      0,
		}

		mockSvc.On("FindDuplicatesByGUID", mock.Anything, userID).Return(result, nil).Once()

		if assert.NoError(t, handler.FindDuplicates(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.Equal(t, float64(0), response["total_duplicates"])
		}
		mockSvc.AssertExpectations(t)
	})
}

