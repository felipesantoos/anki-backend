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

// SharedDeckRatingHandler handles shared deck rating HTTP requests
type SharedDeckRatingHandler struct {
	service primary.ISharedDeckRatingService
}

// NewSharedDeckRatingHandler creates a new SharedDeckRatingHandler instance
func NewSharedDeckRatingHandler(service primary.ISharedDeckRatingService) *SharedDeckRatingHandler {
	return &SharedDeckRatingHandler{
		service: service,
	}
}

// Create handles POST /api/v1/marketplace/ratings
// @Summary Rate a shared deck
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateSharedDeckRatingRequest true "Rating request"
// @Success 201 {object} response.SharedDeckRatingResponse
// @Router /api/v1/marketplace/ratings [post]
func (h *SharedDeckRatingHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateSharedDeckRatingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	r, err := h.service.Create(ctx, userID, req.SharedDeckID, req.Rating, req.Comment)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToSharedDeckRatingResponse(r))
}

// FindBySharedDeckID handles GET /api/v1/marketplace/decks/:id/ratings
// @Summary List ratings for a shared deck
// @Tags marketplace
// @Produce json
// @Param id path int true "Shared Deck ID"
// @Success 200 {array} response.SharedDeckRatingResponse
// @Router /api/v1/marketplace/decks/{id}/ratings [get]
func (h *SharedDeckRatingHandler) FindBySharedDeckID(c echo.Context) error {
	ctx := c.Request().Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	ratings, err := h.service.FindBySharedDeckID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToSharedDeckRatingResponseList(ratings))
}

// Update handles PUT /api/v1/marketplace/decks/:id/ratings
// @Summary Update rating for a shared deck
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shared Deck ID"
// @Param request body request.UpdateSharedDeckRatingRequest true "Update request"
// @Success 200 {object} response.SharedDeckRatingResponse
// @Router /api/v1/marketplace/decks/{id}/ratings [put]
func (h *SharedDeckRatingHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateSharedDeckRatingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	r, err := h.service.Update(ctx, userID, id, req.Rating, req.Comment)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToSharedDeckRatingResponse(r))
}

// Delete handles DELETE /api/v1/marketplace/decks/:id/ratings
// @Summary Remove rating from a shared deck
// @Tags marketplace
// @Security BearerAuth
// @Param id path int true "Shared Deck ID"
// @Success 204 "No Content"
// @Router /api/v1/marketplace/decks/{id}/ratings [delete]
func (h *SharedDeckRatingHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

