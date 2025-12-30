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
	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProfileHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockProfileService)
	handler := handlers.NewProfileHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateProfileRequest{Name: "Personal"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		p, _ := profile.NewBuilder().WithID(1).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Name).Return(p, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestProfileHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockProfileService)
	handler := handlers.NewProfileHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		p1, _ := profile.NewBuilder().WithID(1).WithUserID(userID).WithName("P1").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return([]*profile.Profile{p1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestProfileHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockProfileService)
	handler := handlers.NewProfileHandler(mockSvc)
	userID := int64(1)
	profileID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateProfileRequest{Name: "Updated"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/profiles/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		p, _ := profile.NewBuilder().WithID(profileID).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Update", mock.Anything, userID, profileID, reqBody.Name).Return(p, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestProfileHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockProfileService)
	handler := handlers.NewProfileHandler(mockSvc)
	userID := int64(1)
	profileID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/profiles/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, profileID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

