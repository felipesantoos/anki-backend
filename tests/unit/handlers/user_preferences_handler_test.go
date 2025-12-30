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
	"github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserPreferencesHandler_FindByUserID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockUserPreferencesService)
	handler := handlers.NewUserPreferencesHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		up, _ := userpreferences.NewBuilder().WithID(1).WithUserID(userID).Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return(up, nil).Once()

		if assert.NoError(t, handler.FindByUserID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestUserPreferencesHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockUserPreferencesService)
	handler := handlers.NewUserPreferencesHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateUserPreferencesRequest{Language: "pt-BR"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/user/preferences", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*userpreferences.UserPreferences")).Return(nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestUserPreferencesHandler_ResetToDefaults(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockUserPreferencesService)
	handler := handlers.NewUserPreferencesHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences/reset", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		up, _ := userpreferences.NewBuilder().WithID(1).WithUserID(userID).Build()
		mockSvc.On("ResetToDefaults", mock.Anything, userID).Return(up, nil).Once()

		if assert.NoError(t, handler.ResetToDefaults(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}
