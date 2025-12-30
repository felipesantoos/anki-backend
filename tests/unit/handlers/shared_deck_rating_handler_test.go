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
	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func strPtr(s string) *string {
	return &s
}

func TestSharedDeckRatingHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckRatingService)
	handler := handlers.NewSharedDeckRatingHandler(mockSvc)
	userID := int64(1)
	sharedDeckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateSharedDeckRatingRequest{
			SharedDeckID: sharedDeckID,
			Rating:       5,
			Comment:      strPtr("Great!"),
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/ratings", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		ratingVO, _ := valueobjects.NewSharedDeckRating(reqBody.Rating)
		r, _ := shareddeckrating.NewBuilder().WithID(1).WithUserID(userID).WithSharedDeckID(sharedDeckID).WithRating(ratingVO).Build()
		mockSvc.On("Create", mock.Anything, userID, sharedDeckID, reqBody.Rating, reqBody.Comment).Return(r, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSharedDeckRatingHandler_FindBySharedDeckID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckRatingService)
	handler := handlers.NewSharedDeckRatingHandler(mockSvc)
	sharedDeckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/shared-decks/:id/ratings")
		c.SetParamNames("id")
		c.SetParamValues("10")

		ratingVO, _ := valueobjects.NewSharedDeckRating(5)
		r1, _ := shareddeckrating.NewBuilder().WithID(1).WithUserID(1).WithSharedDeckID(sharedDeckID).WithRating(ratingVO).Build()
		mockSvc.On("FindBySharedDeckID", mock.Anything, sharedDeckID).Return([]*shareddeckrating.SharedDeckRating{r1}, nil).Once()

		if assert.NoError(t, handler.FindBySharedDeckID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSharedDeckRatingHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckRatingService)
	handler := handlers.NewSharedDeckRatingHandler(mockSvc)
	userID := int64(1)
	sharedDeckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateSharedDeckRatingRequest{
			Rating:  4,
			Comment: strPtr("Updated"),
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/shared-decks/:id/ratings")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		ratingVO, _ := valueobjects.NewSharedDeckRating(reqBody.Rating)
		r, _ := shareddeckrating.NewBuilder().WithID(1).WithUserID(userID).WithSharedDeckID(sharedDeckID).WithRating(ratingVO).Build()
		mockSvc.On("Update", mock.Anything, userID, sharedDeckID, reqBody.Rating, mock.Anything).Return(r, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSharedDeckRatingHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSharedDeckRatingService)
	handler := handlers.NewSharedDeckRatingHandler(mockSvc)
	userID := int64(1)
	sharedDeckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/shared-decks/:id/ratings")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, sharedDeckID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}
