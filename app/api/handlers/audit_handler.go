package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// AuditHandler handles audit log related HTTP requests (deletion logs, undo history)
type AuditHandler struct {
	deletionLogService primary.IDeletionLogService
	undoHistoryService primary.IUndoHistoryService
}

// NewAuditHandler creates a new AuditHandler instance
func NewAuditHandler(deletionLogService primary.IDeletionLogService, undoHistoryService primary.IUndoHistoryService) *AuditHandler {
	return &AuditHandler{
		deletionLogService: deletionLogService,
		undoHistoryService: undoHistoryService,
	}
}

// GetDeletionLogs handles GET /api/v1/audit/deletions
// @Summary List deletion logs for current user
// @Tags audit
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.DeletionLogResponse
// @Router /api/v1/audit/deletions [get]
func (h *AuditHandler) GetDeletionLogs(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	logs, err := h.deletionLogService.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToDeletionLogResponseList(logs))
}

// GetUndoHistory handles GET /api/v1/audit/undo
// @Summary List undo history for current user
// @Tags audit
// @Produce json
// @Param limit query int false "Number of records to return" default(10)
// @Security BearerAuth
// @Success 200 {array} response.UndoHistoryResponse
// @Router /api/v1/audit/undo [get]
func (h *AuditHandler) GetUndoHistory(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	history, err := h.undoHistoryService.FindLatest(ctx, userID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToUndoHistoryResponseList(history))
}

// DeleteUndoHistory handles DELETE /api/v1/audit/undo/:id
// @Summary Remove an undo history record
// @Tags audit
// @Security BearerAuth
// @Param id path int true "Undo history record ID"
// @Success 204 "No Content"
// @Router /api/v1/audit/undo/{id} [delete]
func (h *AuditHandler) DeleteUndoHistory(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.undoHistoryService.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

