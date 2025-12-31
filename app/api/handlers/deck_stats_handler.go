package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// DeckStatsHandler handles deck statistics-related HTTP requests
type DeckStatsHandler struct {
	service primary.IDeckStatsService
}

// NewDeckStatsHandler creates a new DeckStatsHandler instance
func NewDeckStatsHandler(service primary.IDeckStatsService) *DeckStatsHandler {
	return &DeckStatsHandler{
		service: service,
	}
}

// GetStats handles GET /api/v1/decks/:id/stats
// @Summary Get deck statistics
// @Description Returns study statistics for a specific deck
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deck ID"
// @Success 200 {object} response.DeckStatsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/decks/{id}/stats [get]
func (h *DeckStatsHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	stats, err := h.service.GetStats(ctx, userID, id)
	if err != nil {
		return err // Custom error handler will map this to 404 if ownership fails
	}

	return c.JSON(http.StatusOK, mappers.ToDeckStatsResponse(stats))
}

