package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// MaintenanceHandler handles maintenance-related HTTP requests
type MaintenanceHandler struct {
	cardService primary.ICardService
}

// NewMaintenanceHandler creates a new MaintenanceHandler instance
func NewMaintenanceHandler(cardService primary.ICardService) *MaintenanceHandler {
	return &MaintenanceHandler{
		cardService: cardService,
	}
}

// GetEmptyCards handles GET /api/v1/maintenance/empty-cards
// @Summary List empty cards
// @Description Find cards where the front template renders to empty
// @Tags maintenance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.EmptyCardsResponse
// @Router /api/v1/maintenance/empty-cards [get]
func (h *MaintenanceHandler) GetEmptyCards(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	cards, err := h.cardService.FindEmptyCards(ctx, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.EmptyCardsResponse{
		Count: len(cards),
		Data:  mappers.ToCardResponseList(cards),
	})
}

// CleanupEmptyCards handles POST /api/v1/maintenance/empty-cards/cleanup
// @Summary Cleanup empty cards
// @Description Delete all cards where the front template renders to empty
// @Tags maintenance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.CleanupEmptyCardsResponse
// @Router /api/v1/maintenance/empty-cards/cleanup [post]
func (h *MaintenanceHandler) CleanupEmptyCards(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	count, err := h.cardService.CleanupEmptyCards(ctx, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.CleanupEmptyCardsResponse{
		DeletedCount: count,
	})
}
