package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	handler := handlers.NewNoteHandler(mockSvc)
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
	handler := handlers.NewNoteHandler(mockSvc)
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

