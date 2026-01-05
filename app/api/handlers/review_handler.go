package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// ReviewHandler handles card review-related HTTP requests
type ReviewHandler struct {
	service primary.IReviewService
}

// NewReviewHandler creates a new ReviewHandler instance
func NewReviewHandler(service primary.IReviewService) *ReviewHandler {
	return &ReviewHandler{
		service: service,
	}
}

// Create handles POST /api/v1/reviews
// @Summary Record a card review
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateReviewRequest true "Review request"
// @Success 201 {object} response.ReviewResponse
// @Router /api/v1/reviews [post]
func (h *ReviewHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	review, err := h.service.Create(ctx, userID, req.CardID, req.Rating, req.TimeMs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToReviewResponse(review))
}

// FindByCardID handles GET /api/v1/cards/:cardID/reviews
// @Summary List reviews for a card
// @Tags reviews
// @Produce json
// @Security BearerAuth
// @Param cardID path int true "Card ID"
// @Success 200 {array} response.ReviewResponse
// @Router /api/v1/cards/{cardID}/reviews [get]
func (h *ReviewHandler) FindByCardID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	cardID, _ := strconv.ParseInt(c.Param("cardID"), 10, 64)

	reviews, err := h.service.FindByCardID(ctx, userID, cardID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToReviewResponseList(reviews))
}

