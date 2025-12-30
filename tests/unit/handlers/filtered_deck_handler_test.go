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
	"github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilteredDeckHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockFilteredDeckService)
	handler := handlers.NewFilteredDeckHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateFilteredDeckRequest{
			Name:         "Filtered",
			SearchFilter: "is:due",
			Limit:        100,
			OrderBy:      "random",
			Reschedule:   true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/filtered-decks", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		fd, _ := filtereddeck.NewBuilder().WithID(1).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Name, reqBody.SearchFilter, reqBody.Limit, reqBody.OrderBy, reqBody.Reschedule).Return(fd, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestFilteredDeckHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockFilteredDeckService)
	handler := handlers.NewFilteredDeckHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/filtered-decks", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		fd1, _ := filtereddeck.NewBuilder().WithID(1).WithUserID(userID).WithName("FD1").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return([]*filtereddeck.FilteredDeck{fd1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestFilteredDeckHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockFilteredDeckService)
	handler := handlers.NewFilteredDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateFilteredDeckRequest{
			Name:         "Updated",
			SearchFilter: "is:new",
			Limit:        50,
			OrderBy:      "added",
			Reschedule:   false,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/filtered-decks/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		fd, _ := filtereddeck.NewBuilder().WithID(deckID).WithUserID(userID).WithName(reqBody.Name).Build()
		mockSvc.On("Update", mock.Anything, userID, deckID, reqBody.Name, reqBody.SearchFilter, reqBody.Limit, reqBody.OrderBy, reqBody.Reschedule).Return(fd, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestFilteredDeckHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockFilteredDeckService)
	handler := handlers.NewFilteredDeckHandler(mockSvc)
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/filtered-decks/:id")
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

