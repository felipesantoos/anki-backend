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
	"github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddOnHandler_Install(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockAddOnService)
	handler := handlers.NewAddOnHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.InstallAddOnRequest{
			Code:    "123456",
			Name:    "Test AddOn",
			Version: "1.0.0",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/addons", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		a, _ := addon.NewBuilder().WithID(1).WithUserID(userID).WithCode(reqBody.Code).WithName(reqBody.Name).Build()
		mockSvc.On("Install", mock.Anything, userID, reqBody.Code, reqBody.Name, reqBody.Version, mock.Anything).Return(a, nil).Once()

		if assert.NoError(t, handler.Install(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestAddOnHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockAddOnService)
	handler := handlers.NewAddOnHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/addons", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		a1, _ := addon.NewBuilder().WithID(1).WithUserID(userID).WithCode("C1").WithName("A1").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return([]*addon.AddOn{a1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestAddOnHandler_UpdateConfig(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockAddOnService)
	handler := handlers.NewAddOnHandler(mockSvc)
	userID := int64(1)
	code := "123456"

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateAddOnConfigRequest{ConfigJSON: "{}"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/addons/:code/config")
		c.SetParamNames("code")
		c.SetParamValues(code)
		c.Set(middlewares.UserIDContextKey, userID)

		a, _ := addon.NewBuilder().WithID(1).WithUserID(userID).WithCode(code).WithName("Test").Build()
		mockSvc.On("UpdateConfig", mock.Anything, userID, code, reqBody.ConfigJSON).Return(a, nil).Once()

		if assert.NoError(t, handler.UpdateConfig(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestAddOnHandler_Uninstall(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockAddOnService)
	handler := handlers.NewAddOnHandler(mockSvc)
	userID := int64(1)
	code := "123456"

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/addons/:code")
		c.SetParamNames("code")
		c.SetParamValues(code)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Uninstall", mock.Anything, userID, code).Return(nil).Once()

		if assert.NoError(t, handler.Uninstall(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

