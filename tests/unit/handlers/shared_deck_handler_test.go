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
	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSharedDeckHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckService)
	handler := handlers.NewSharedDeckHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateSharedDeckRequest{
			Name:        "Shared Deck",
			PackagePath: "/path",
			PackageSize: 1024,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shared-decks", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		sd, _ := shareddeck.NewBuilder().WithID(1).WithAuthorID(userID).WithName(reqBody.Name).WithPackagePath(reqBody.PackagePath).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Name, mock.Anything, mock.Anything, reqBody.PackagePath, reqBody.PackageSize, mock.Anything).Return(sd, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSharedDeckHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckService)
	handler := handlers.NewSharedDeckHandler(mockSvc)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shared-decks", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		sd1, _ := shareddeck.NewBuilder().WithID(1).WithAuthorID(1).WithName("SD1").WithPackagePath("/p1").Build()
		mockSvc.On("FindAll", mock.Anything, mock.Anything, mock.Anything).Return([]*shareddeck.SharedDeck{sd1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSharedDeckHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckService)
	handler := handlers.NewSharedDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateSharedDeckRequest{
			Name:     "Updated",
			IsPublic: true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/shared-decks/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		sd, _ := shareddeck.NewBuilder().WithID(deckID).WithAuthorID(userID).WithName(reqBody.Name).WithPackagePath("/p").Build()
		mockSvc.On("Update", mock.Anything, userID, deckID, reqBody.Name, mock.Anything, mock.Anything, reqBody.IsPublic, mock.Anything).Return(sd, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSharedDeckHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckService)
	handler := handlers.NewSharedDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/shared-decks/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, deckID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

