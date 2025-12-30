package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// SyncMetaHandler handles synchronization metadata HTTP requests
type SyncMetaHandler struct {
	service primary.ISyncMetaService
}

// NewSyncMetaHandler creates a new SyncMetaHandler instance
func NewSyncMetaHandler(service primary.ISyncMetaService) *SyncMetaHandler {
	return &SyncMetaHandler{
		service: service,
	}
}

// FindMe handles GET /api/v1/sync/meta
// @Summary Get sync metadata for current user
// @Tags sync
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SyncMetaResponse
// @Router /api/v1/sync/meta [get]
func (h *SyncMetaHandler) FindMe(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	sm, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToSyncMetaResponse(sm))
}

// Update handles PUT /api/v1/sync/meta
// @Summary Update sync metadata after synchronization
// @Tags sync
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateSyncMetaRequest true "Update request"
// @Success 200 {object} response.SyncMetaResponse
// @Router /api/v1/sync/meta [put]
func (h *SyncMetaHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.UpdateSyncMetaRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	sm, err := h.service.Update(ctx, userID, req.ClientID, req.LastSyncUSN)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToSyncMetaResponse(sm))
}

