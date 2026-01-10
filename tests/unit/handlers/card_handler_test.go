package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/pkg/ownership"
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

	t.Run("Invalid ID format (non-numeric)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("invalid")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.FindByID(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Contains(t, httpErr.Message.(string), "Invalid card ID format")
		}
		// Service should not be called when validation fails
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.FindByID(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Contains(t, httpErr.Message.(string), "Card ID must be greater than 0")
		}
		// Service should not be called when validation fails
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID (negative)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("-1")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.FindByID(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Contains(t, httpErr.Message.(string), "Card ID must be greater than 0")
		}
		// Service should not be called when validation fails
		mockSvc.AssertExpectations(t)
	})

	t.Run("Card not found (404)", func(t *testing.T) {
		notFoundMockSvc := new(MockCardService)
		notFoundHandler := handlers.NewCardHandler(notFoundMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set(middlewares.UserIDContextKey, userID)

		notFoundMockSvc.On("FindByID", mock.Anything, userID, int64(999)).Return(nil, ownership.ErrResourceNotFound).Once()

		err := notFoundHandler.FindByID(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Contains(t, httpErr.Message.(string), "Card not found")
		}
		notFoundMockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
		serviceErrorMockSvc := new(MockCardService)
		serviceErrorHandler := handlers.NewCardHandler(serviceErrorMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		serviceErrorMockSvc.On("FindByID", mock.Anything, userID, cardID).Return(nil, echo.NewHTTPError(http.StatusInternalServerError, "database error")).Once()

		err := serviceErrorHandler.FindByID(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
		}
		serviceErrorMockSvc.AssertExpectations(t)
	})

	t.Run("Cross-user isolation (returns 404)", func(t *testing.T) {
		crossUserMockSvc := new(MockCardService)
		crossUserHandler := handlers.NewCardHandler(crossUserMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		// User tries to access another user's card - should return 404
		crossUserMockSvc.On("FindByID", mock.Anything, userID, cardID).Return(nil, ownership.ErrResourceNotFound).Once()

		err := crossUserHandler.FindByID(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Contains(t, httpErr.Message.(string), "Card not found")
		}
		crossUserMockSvc.AssertExpectations(t)
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

	t.Run("Invalid ID format (non-numeric)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("invalid")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Suspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Suspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (negative)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("-1")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Suspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Card not found (404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Suspend", mock.Anything, userID, int64(999)).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Suspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Suspend", mock.Anything, userID, cardID).Return(errors.New("database error")).Once()

		err := handler.Suspend(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Cross-user isolation (returns 404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/suspend")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, int64(99)) // Different user ID

		mockSvc.On("Suspend", mock.Anything, int64(99), cardID).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Suspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_Unsuspend(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Unsuspend", mock.Anything, userID, cardID).Return(nil).Once()

		if assert.NoError(t, handler.Unsuspend(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID format (non-numeric)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("invalid")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Unsuspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Unsuspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (negative)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("-1")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Unsuspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Card not found (404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Unsuspend", mock.Anything, userID, int64(999)).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Unsuspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Unsuspend", mock.Anything, userID, cardID).Return(errors.New("database error")).Once()

		err := handler.Unsuspend(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Cross-user isolation (returns 404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unsuspend")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, int64(99)) // Different user ID

		mockSvc.On("Unsuspend", mock.Anything, int64(99), cardID).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Unsuspend(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_SetFlag(t *testing.T) {
	e := echo.New()
	e.Validator = middlewares.NewCustomValidator()
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

	t.Run("Invalid ID format (non-numeric)", func(t *testing.T) {
		reqBody := request.SetCardFlagRequest{Flag: 1}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("invalid")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.SetFlag(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		reqBody := request.SetCardFlagRequest{Flag: 1}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.SetFlag(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (negative)", func(t *testing.T) {
		reqBody := request.SetCardFlagRequest{Flag: 1}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("-1")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.SetFlag(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Card not found (404)", func(t *testing.T) {
		reqBody := request.SetCardFlagRequest{Flag: 1}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("SetFlag", mock.Anything, userID, int64(999), 1).Return(ownership.ErrResourceNotFound).Once()

		err := handler.SetFlag(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid flag value (domain error)", func(t *testing.T) {
		// Use flag 9 to bypass DTO validation (which allows 0-7) and test domain error
		// Actually, DTO validation will catch flag=9, so let's use a flag that passes DTO validation
		// but would fail domain validation. Since DTO already validates 0-7, we need to mock
		// a scenario where the service returns ErrInvalidFlag anyway (edge case)
		reqBody := request.SetCardFlagRequest{Flag: 7} // Valid DTO value
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		// Simulate service returning domain error (edge case where domain validation might fail)
		mockSvc.On("SetFlag", mock.Anything, userID, cardID, 7).Return(card.ErrInvalidFlag).Once()

		err := handler.SetFlag(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Flag must be between 0 and 7", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
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

		mockSvc.On("SetFlag", mock.Anything, userID, cardID, 1).Return(errors.New("database error")).Once()

		err := handler.SetFlag(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Cross-user isolation (returns 404)", func(t *testing.T) {
		reqBody := request.SetCardFlagRequest{Flag: 1}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/flag")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, int64(99)) // Different user ID

		mockSvc.On("SetFlag", mock.Anything, int64(99), cardID, 1).Return(ownership.ErrResourceNotFound).Once()

		err := handler.SetFlag(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_Bury(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Bury", mock.Anything, userID, cardID).Return(nil).Once()

		if assert.NoError(t, handler.Bury(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID format (non-numeric)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("invalid")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Bury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Bury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (negative)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("-1")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Bury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Card not found (404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Bury", mock.Anything, userID, int64(999)).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Bury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Bury", mock.Anything, userID, cardID).Return(errors.New("database error")).Once()

		err := handler.Bury(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Cross-user isolation (returns 404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/bury")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, int64(99)) // Different user ID

		mockSvc.On("Bury", mock.Anything, int64(99), cardID).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Bury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_Unbury(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Unbury", mock.Anything, userID, cardID).Return(nil).Once()

		if assert.NoError(t, handler.Unbury(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID format (non-numeric)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("invalid")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Unbury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Unbury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Invalid ID (negative)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("-1")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Unbury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
		mockSvc.AssertExpectations(t) // Service should not be called
	})

	t.Run("Card not found (404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Unbury", mock.Anything, userID, int64(999)).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Unbury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Unbury", mock.Anything, userID, cardID).Return(errors.New("database error")).Once()

		err := handler.Unbury(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Cross-user isolation (returns 404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/unbury")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, int64(99)) // Different user ID

		mockSvc.On("Unbury", mock.Anything, int64(99), cardID).Return(ownership.ErrResourceNotFound).Once()

		err := handler.Unbury(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_FindAll(t *testing.T) {
	e := echo.New()
	e.Validator = middlewares.NewCustomValidator()
	mockSvc := new(MockCardService)
	handler := handlers.NewCardHandler(mockSvc)
	userID := int64(1)

	t.Run("Success with no filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card { c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build(); return c }(),
			func() *card.Card { c, _ := card.NewBuilder().WithID(2).WithNoteID(2).WithDeckID(1).Build(); return c }(),
		}
		total := 2

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.Limit == 20 && filters.Offset == 0 &&
				filters.DeckID == nil && filters.State == nil && filters.Flag == nil &&
				filters.Suspended == nil && filters.Buried == nil
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Len(t, res.Data, 2)
			assert.Equal(t, 1, res.Pagination.Page)
			assert.Equal(t, 20, res.Pagination.Limit)
			assert.Equal(t, total, res.Pagination.Total)
			assert.Equal(t, 1, res.Pagination.TotalPages)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with deck_id filter", func(t *testing.T) {
		deckID := int64(10)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?deck_id=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(deckID).Build()
				return c
			}(),
		}
		total := 1

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.DeckID != nil && *filters.DeckID == deckID
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Len(t, res.Data, 1)
			assert.Equal(t, deckID, res.Data[0].DeckID)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with state filter", func(t *testing.T) {
		state := "new"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?state=new", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).WithState(valueobjects.CardStateNew).Build()
				return c
			}(),
		}
		total := 1

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.State != nil && *filters.State == state
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Len(t, res.Data, 1)
			assert.Equal(t, state, res.Data[0].State)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with flag filter", func(t *testing.T) {
		flag := 3
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?flag=3", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build()
				c.SetFlag(flag)
				return c
			}(),
		}
		total := 1

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.Flag != nil && *filters.Flag == flag
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Len(t, res.Data, 1)
			assert.Equal(t, flag, res.Data[0].Flags)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with suspended filter", func(t *testing.T) {
		suspended := true
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?suspended=true", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build()
				c.Suspend()
				return c
			}(),
		}
		total := 1

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.Suspended != nil && *filters.Suspended == suspended
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Len(t, res.Data, 1)
			assert.True(t, res.Data[0].Suspended)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with buried filter", func(t *testing.T) {
		buried := true
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?buried=true", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build()
				c.Bury()
				return c
			}(),
		}
		total := 1

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.Buried != nil && *filters.Buried == buried
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?page=2&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card { c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build(); return c }(),
		}
		total := 25

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.Limit == 10 && filters.Offset == 10
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Equal(t, 2, res.Pagination.Page)
			assert.Equal(t, 10, res.Pagination.Limit)
			assert.Equal(t, total, res.Pagination.Total)
			assert.Equal(t, 3, res.Pagination.TotalPages)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success with multiple filters", func(t *testing.T) {
		deckID := int64(10)
		state := "review"
		flag := 2
		suspended := false
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?deck_id=10&state=review&flag=2&suspended=false", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(deckID).WithState(valueobjects.CardStateReview).Build()
				c.SetFlag(flag)
				return c
			}(),
		}
		total := 1

		mockSvc.On("FindAll", mock.Anything, userID, mock.MatchedBy(func(filters card.CardFilters) bool {
			return filters.DeckID != nil && *filters.DeckID == deckID &&
				filters.State != nil && *filters.State == state &&
				filters.Flag != nil && *filters.Flag == flag &&
				filters.Suspended != nil && *filters.Suspended == suspended
		})).Return(cards, total, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Validation error - invalid state", func(t *testing.T) {
		validationMockSvc := new(MockCardService)
		validationHandler := handlers.NewCardHandler(validationMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?state=invalid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		err := validationHandler.FindAll(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		}
		// Service should not be called when validation fails
		validationMockSvc.AssertExpectations(t)
	})

	t.Run("Validation error - invalid flag", func(t *testing.T) {
		validationMockSvc := new(MockCardService)
		validationHandler := handlers.NewCardHandler(validationMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?flag=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		err := validationHandler.FindAll(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		}
		// Service should not be called when validation fails
		validationMockSvc.AssertExpectations(t)
	})

	t.Run("Validation error - invalid limit (exceeds max)", func(t *testing.T) {
		validationMockSvc := new(MockCardService)
		validationHandler := handlers.NewCardHandler(validationMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?limit=101", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		err := validationHandler.FindAll(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		}
		// Service should not be called when validation fails
		validationMockSvc.AssertExpectations(t)
	})

	t.Run("Service error handling", func(t *testing.T) {
		serviceErrorMockSvc := new(MockCardService)
		serviceErrorHandler := handlers.NewCardHandler(serviceErrorMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		serviceErrorMockSvc.On("FindAll", mock.Anything, userID, mock.Anything).Return(nil, 0, echo.NewHTTPError(http.StatusInternalServerError, "database error")).Once()

		err := serviceErrorHandler.FindAll(c)

		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
		}
		serviceErrorMockSvc.AssertExpectations(t)
	})

	t.Run("Pagination with zero total", func(t *testing.T) {
		zeroTotalMockSvc := new(MockCardService)
		zeroTotalHandler := handlers.NewCardHandler(zeroTotalMockSvc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards")
		c.Set(middlewares.UserIDContextKey, userID)

		cards := []*card.Card{}
		total := 0

		zeroTotalMockSvc.On("FindAll", mock.Anything, userID, mock.Anything).Return(cards, total, nil).Once()

		if assert.NoError(t, zeroTotalHandler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Equal(t, 0, res.Pagination.Total)
			assert.Equal(t, 1, res.Pagination.TotalPages) // Should be 1 even with 0 total
		}
		zeroTotalMockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_FindLeeches(t *testing.T) {
	e := echo.New()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		cards := []*card.Card{
			func() *card.Card { c, _ := card.NewBuilder().WithID(1).WithLapses(10).Build(); return c }(),
		}
		mockSvc.On("FindLeeches", mock.Anything, userID, 20, 0).Return(cards, 1, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/leeches", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		if assert.NoError(t, handler.FindLeeches(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.ListCardsResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Len(t, res.Data, 1)
			assert.Equal(t, 1, res.Pagination.Total)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Pagination", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		mockSvc.On("FindLeeches", mock.Anything, userID, 10, 10).Return([]*card.Card{}, 0, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/leeches?page=2&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		if assert.NoError(t, handler.FindLeeches(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		mockSvc.On("FindLeeches", mock.Anything, userID, 20, 0).Return(nil, 0, errors.New("service error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/leeches", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.FindLeeches(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_Reset(t *testing.T) {
	e := echo.New()
	e.Validator = middlewares.NewCustomValidator()
	userID := int64(1)
	cardID := int64(123)

	t.Run("Success - type new", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		resetReq := request.ResetCardRequest{Type: "new"}
		b, _ := json.Marshal(resetReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/reset", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/reset")
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Reset", mock.Anything, userID, cardID, "new").Return(nil).Once()

		if assert.NoError(t, handler.Reset(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Success - type forget", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		resetReq := request.ResetCardRequest{Type: "forget"}
		b, _ := json.Marshal(resetReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/reset", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/reset")
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Reset", mock.Anything, userID, cardID, "forget").Return(nil).Once()

		if assert.NoError(t, handler.Reset(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/abc/reset", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("abc")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Reset(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/0/reset", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Reset(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
	})

	t.Run("Validation error - invalid reset type", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		resetReq := map[string]string{"type": "invalid"}
		b, _ := json.Marshal(resetReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/reset", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Reset(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		}
	})

	t.Run("Card not found", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		resetReq := request.ResetCardRequest{Type: "new"}
		b, _ := json.Marshal(resetReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/reset", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Reset", mock.Anything, userID, cardID, "new").Return(ownership.ErrResourceNotFound).Once()

		err := handler.Reset(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
	})

	t.Run("Service error", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		resetReq := request.ResetCardRequest{Type: "new"}
		b, _ := json.Marshal(resetReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/reset", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Reset", mock.Anything, userID, cardID, "new").Return(errors.New("service error")).Once()

		err := handler.Reset(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_SetDueDate(t *testing.T) {
	e := echo.New()
	e.Validator = middlewares.NewCustomValidator()
	userID := int64(1)
	cardID := int64(123)
	due := int64(1705324200000)

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		dueReq := request.SetCardDueDateRequest{Due: due}
		b, _ := json.Marshal(dueReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/due", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/due")
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("SetDueDate", mock.Anything, userID, cardID, due).Return(nil).Once()

		if assert.NoError(t, handler.SetDueDate(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/abc/due", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("abc")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.SetDueDate(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/0/due", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.SetDueDate(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
	})

	t.Run("Validation error - negative due date", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		dueReq := map[string]interface{}{"due": -1}
		b, _ := json.Marshal(dueReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/due", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.SetDueDate(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		}
	})

	t.Run("Card not found", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		dueReq := request.SetCardDueDateRequest{Due: due}
		b, _ := json.Marshal(dueReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/due", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("SetDueDate", mock.Anything, userID, cardID, due).Return(ownership.ErrResourceNotFound).Once()

		err := handler.SetDueDate(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
	})

	t.Run("Service error", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		dueReq := request.SetCardDueDateRequest{Due: due}
		b, _ := json.Marshal(dueReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/123/due", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("SetDueDate", mock.Anything, userID, cardID, due).Return(errors.New("service error")).Once()

		err := handler.SetDueDate(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_Reposition(t *testing.T) {
	e := echo.New()
	e.Validator = middlewares.NewCustomValidator()
	userID := int64(1)
	cardIDs := []int64{1, 2, 3}

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		repoReq := request.RepositionCardsRequest{
			CardIDs: cardIDs,
			Start:   10,
			Step:    2,
			Shift:   true,
		}
		b, _ := json.Marshal(repoReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/reposition", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Reposition", mock.Anything, userID, cardIDs, 10, 2, true).Return(nil).Once()

		if assert.NoError(t, handler.Reposition(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Validation Error - empty IDs", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		repoReq := request.RepositionCardsRequest{
			CardIDs: []int64{},
			Start:   0,
			Step:    1,
		}
		b, _ := json.Marshal(repoReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/reposition", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Reposition(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		}
	})

	t.Run("Validation Error - negative start", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		repoReq := request.RepositionCardsRequest{
			CardIDs: cardIDs,
			Start:   -1,
			Step:    1,
		}
		b, _ := json.Marshal(repoReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/reposition", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.Reposition(c)
		assert.Error(t, err)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		repoReq := request.RepositionCardsRequest{
			CardIDs: cardIDs,
			Start:   0,
			Step:    1,
		}
		b, _ := json.Marshal(repoReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/reposition", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Reposition", mock.Anything, userID, cardIDs, 0, 1, false).Return(errors.New("service error")).Once()

		err := handler.Reposition(c)
		assert.Error(t, err)
		mockSvc.AssertExpectations(t)
	})
}

func TestCardHandler_GetPosition(t *testing.T) {
	e := echo.New()
	userID := int64(1)
	cardID := int64(123)

	t.Run("Success", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		mockSvc.On("GetPosition", mock.Anything, userID, cardID).Return(100, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/123/position", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:id/position")
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		if assert.NoError(t, handler.GetPosition(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var res response.CardPositionResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Equal(t, 100, res.Position)
		}
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/abc/position", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("abc")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.GetPosition(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid card ID format", httpErr.Message)
		}
	})

	t.Run("Invalid ID (zero)", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/0/position", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("0")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.GetPosition(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Card ID must be greater than 0", httpErr.Message)
		}
	})

	t.Run("Card not found", func(t *testing.T) {
		mockSvc := new(MockCardService)
		handler := handlers.NewCardHandler(mockSvc)

		mockSvc.On("GetPosition", mock.Anything, userID, cardID).Return(0, ownership.ErrResourceNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/123/position", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("123")
		c.Set(middlewares.UserIDContextKey, userID)

		err := handler.GetPosition(c)
		assert.Error(t, err)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			assert.Equal(t, "Card not found", httpErr.Message)
		}
	})
}
