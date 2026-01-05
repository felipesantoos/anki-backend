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
	"github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteTypeHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteTypeService)
	handler := handlers.NewNoteTypeHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateNoteTypeRequest{
			Name:          "Basic",
			FieldsJSON:    "[]",
			CardTypesJSON: "[]",
			TemplatesJSON: "{}",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/note-types", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		nt, _ := notetype.NewBuilder().WithID(1).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Name, reqBody.FieldsJSON, reqBody.CardTypesJSON, reqBody.TemplatesJSON).Return(nt, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		e := echo.New()
		e.Validator = middlewares.NewCustomValidator()
		mockSvc := new(MockNoteTypeService)
		handler := handlers.NewNoteTypeHandler(mockSvc)

		tests := []struct {
			name    string
			reqBody map[string]interface{}
			wantCode int
		}{
			{
				name: "missing name",
				reqBody: map[string]interface{}{
					"fields_json":      `[{"name": "Front"}]`,
					"card_types_json":  `[{"name": "Card 1"}]`,
					"templates_json":   `[]`,
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name: "missing fields_json",
				reqBody: map[string]interface{}{
					"name":            "Test",
					"card_types_json": `[{"name": "Card 1"}]`,
					"templates_json":  `[]`,
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name: "missing card_types_json",
				reqBody: map[string]interface{}{
					"name":           "Test",
					"fields_json":    `[{"name": "Front"}]`,
					"templates_json": `[]`,
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name: "missing templates_json",
				reqBody: map[string]interface{}{
					"name":           "Test",
					"fields_json":    `[{"name": "Front"}]`,
					"card_types_json": `[{"name": "Card 1"}]`,
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonBody, _ := json.Marshal(tt.reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/note-types", bytes.NewReader(jsonBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set(middlewares.UserIDContextKey, userID)

				err := handler.Create(c)

				if err == nil {
					t.Fatalf("Create() expected error, got nil")
				}

				if httpErr, ok := err.(*echo.HTTPError); ok {
					if httpErr.Code != tt.wantCode {
						t.Errorf("Create() status code = %d, want %d", httpErr.Code, tt.wantCode)
					}
				} else {
					t.Errorf("Create() error type = %T, want *echo.HTTPError", err)
				}
			})
		}
	})
}

func TestNoteTypeHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteTypeService)
	handler := handlers.NewNoteTypeHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/note-types", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		nt1, _ := notetype.NewBuilder().WithID(1).WithUserID(userID).WithName("Basic").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID, "").Return([]*notetype.NoteType{nt1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestNoteTypeHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteTypeService)
	handler := handlers.NewNoteTypeHandler(mockSvc)
	userID := int64(1)
	noteTypeID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateNoteTypeRequest{
			Name:          "Updated",
			FieldsJSON:    "[]",
			CardTypesJSON: "[]",
			TemplatesJSON: "{}",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/note-types/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		nt, _ := notetype.NewBuilder().WithID(noteTypeID).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Update", mock.Anything, userID, noteTypeID, reqBody.Name, reqBody.FieldsJSON, reqBody.CardTypesJSON, reqBody.TemplatesJSON).Return(nt, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestNoteTypeHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockNoteTypeService)
	handler := handlers.NewNoteTypeHandler(mockSvc)
	userID := int64(1)
	noteTypeID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/note-types/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, noteTypeID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

