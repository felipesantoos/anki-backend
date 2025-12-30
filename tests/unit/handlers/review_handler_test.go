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
	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReviewHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockReviewService)
	handler := handlers.NewReviewHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateReviewRequest{
			CardID: 10,
			Rating: 3,
			TimeMs: 5000,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		r, _ := review.NewBuilder().WithID(1).WithCardID(reqBody.CardID).WithRating(reqBody.Rating).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.CardID, reqBody.Rating, reqBody.TimeMs).Return(r, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestReviewHandler_FindByCardID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockReviewService)
	handler := handlers.NewReviewHandler(mockSvc)
	userID := int64(1)
	cardID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/cards/:cardID/reviews")
		c.SetParamNames("cardID")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		r1, _ := review.NewBuilder().WithID(1).WithCardID(cardID).WithRating(3).Build()
		mockSvc.On("FindByCardID", mock.Anything, userID, cardID).Return([]*review.Review{r1}, nil).Once()

		if assert.NoError(t, handler.FindByCardID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

