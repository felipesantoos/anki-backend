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
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_GetMe(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockUserService)
	handler := handlers.NewUserHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		email, _ := valueobjects.NewEmail("user@example.com")
		password, _ := valueobjects.NewPassword("password123")
		u, _ := user.NewBuilder().WithID(userID).WithEmail(email).WithPasswordHash(password).Build()
		mockSvc.On("FindByID", mock.Anything, userID).Return(u, nil).Once()

		if assert.NoError(t, handler.GetMe(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestUserHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockUserService)
	handler := handlers.NewUserHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateUserRequest{Email: "new@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/user/me", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		email, _ := valueobjects.NewEmail(reqBody.Email)
		password, _ := valueobjects.NewPassword("password123")
		u, _ := user.NewBuilder().WithID(userID).WithEmail(email).WithPasswordHash(password).Build()
		mockSvc.On("Update", mock.Anything, userID, reqBody.Email).Return(u, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockUserService)
	handler := handlers.NewUserHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}
