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
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCardHandler_FindByID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		cd := &card.Card{}
		cd.SetID(cardID)
		mockSvc.On("FindByID", mock.Anything, userID, cardID).Return(cd, nil).Once()

		if assert.NoError(t, handler.FindByID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_FindByDeckID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/decks/:deckID/cards")
		c.SetParamNames("deckID")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		cd := &card.Card{}
		cd.SetID(1)
		mockSvc.On("FindByDeckID", mock.Anything, userID, deckID).Return([]*card.Card{cd}, nil).Once()

		if assert.NoError(t, handler.FindByDeckID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_Suspend(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Suspend", mock.Anything, userID, cardID).Return(nil).Once()

		if assert.NoError(t, handler.Suspend(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_SetFlag(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.SetCardFlagRequest{Flag: 1}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("SetFlag", mock.Anything, userID, cardID, 1).Return(nil).Once()

		if assert.NoError(t, handler.SetFlag(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

