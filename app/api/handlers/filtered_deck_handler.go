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

// FilteredDeckHandler handles filtered deck-related HTTP requests
type FilteredDeckHandler struct {
	service primary.IFilteredDeckService
}

// NewFilteredDeckHandler creates a new FilteredDeckHandler instance
func NewFilteredDeckHandler(service primary.IFilteredDeckService) *FilteredDeckHandler {
	return &FilteredDeckHandler{
		service: service,
	}
}

// Create handles POST /api/v1/filtered-decks
// @Summary Create a filtered deck
// @Tags filtered-decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateFilteredDeckRequest true "Filtered deck creation request"
// @Success 201 {object} response.FilteredDeckResponse
// @Router /api/v1/filtered-decks [post]
func (h *FilteredDeckHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateFilteredDeckRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	fd, err := h.service.Create(ctx, userID, req.Name, req.SearchFilter, req.Limit, req.OrderBy, req.Reschedule)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToFilteredDeckResponse(fd))
}

// FindAll handles GET /api/v1/filtered-decks
// @Summary List filtered decks
// @Tags filtered-decks
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.FilteredDeckResponse
// @Router /api/v1/filtered-decks [get]
func (h *FilteredDeckHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	decks, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToFilteredDeckResponseList(decks))
}

// Update handles PUT /api/v1/filtered-decks/:id
// @Summary Update filtered deck
// @Tags filtered-decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Filtered Deck ID"
// @Param request body request.UpdateFilteredDeckRequest true "Update request"
// @Success 200 {object} response.FilteredDeckResponse
// @Router /api/v1/filtered-decks/{id} [put]
func (h *FilteredDeckHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateFilteredDeckRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	fd, err := h.service.Update(ctx, userID, id, req.Name, req.SearchFilter, req.Limit, req.OrderBy, req.Reschedule)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToFilteredDeckResponse(fd))
}

// Delete handles DELETE /api/v1/filtered-decks/:id
// @Summary Delete filtered deck
// @Tags filtered-decks
// @Security BearerAuth
// @Param id path int true "Filtered Deck ID"
// @Success 204 "No Content"
// @Router /api/v1/filtered-decks/{id} [delete]
func (h *FilteredDeckHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

