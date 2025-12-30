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

// BackupHandler handles backup-related HTTP requests
type BackupHandler struct {
	service primary.IBackupService
}

// NewBackupHandler creates a new BackupHandler instance
func NewBackupHandler(service primary.IBackupService) *BackupHandler {
	return &BackupHandler{
		service: service,
	}
}

// Create handles POST /api/v1/backups
// @Summary Record a backup
// @Tags backups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateBackupRequest true "Backup creation request"
// @Success 201 {object} response.BackupResponse
// @Router /api/v1/backups [post]
func (h *BackupHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateBackupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	b, err := h.service.Create(ctx, userID, req.Filename, req.Size, req.StoragePath, req.BackupType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToBackupResponse(b))
}

// FindAll handles GET /api/v1/backups
// @Summary List backups
// @Tags backups
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.BackupResponse
// @Router /api/v1/backups [get]
func (h *BackupHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	backups, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToBackupResponseList(backups))
}

// Delete handles DELETE /api/v1/backups/:id
// @Summary Delete backup record
// @Tags backups
// @Security BearerAuth
// @Param id path int true "Backup ID"
// @Success 204 "No Content"
// @Router /api/v1/backups/{id} [delete]
func (h *BackupHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

