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
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeckHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockDeckService)
	handler := handlers.NewDeckHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateDeckRequest{
			Name: "Test Deck",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		d, _ := deck.NewBuilder().WithID(1).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Name, mock.Anything, mock.Anything).Return(d, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			var res map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Equal(t, float64(1), res["id"])
			assert.Equal(t, reqBody.Name, res["name"])
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestDeckHandler_FindByID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockDeckService)
	handler := handlers.NewDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/decks/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Test").Build()
		mockSvc.On("FindByID", mock.Anything, userID, deckID).Return(d, nil).Once()

		if assert.NoError(t, handler.FindByID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestDeckHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockDeckService)
	handler := handlers.NewDeckHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/decks", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		d1, _ := deck.NewBuilder().WithID(1).WithUserID(userID).WithName("D1").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return([]*deck.Deck{d1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestDeckHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockDeckService)
	handler := handlers.NewDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateDeckRequest{
			Name: "Updated Name",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/decks/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Update", mock.Anything, userID, deckID, reqBody.Name, mock.Anything, mock.Anything).Return(d, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestDeckHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockDeckService)
	handler := handlers.NewDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/decks/:id")
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

