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

// MediaHandler handles media file-related HTTP requests
type MediaHandler struct {
	service primary.IMediaService
}

// NewMediaHandler creates a new MediaHandler instance
func NewMediaHandler(service primary.IMediaService) *MediaHandler {
	return &MediaHandler{
		service: service,
	}
}

// Create handles POST /api/v1/media
// @Summary Record a media file
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateMediaRequest true "Media creation request"
// @Success 201 {object} response.MediaResponse
// @Router /api/v1/media [post]
func (h *MediaHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateMediaRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	m, err := h.service.Create(ctx, userID, req.Filename, req.Hash, req.Size, req.MimeType, req.StoragePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToMediaResponse(m))
}

// FindByID handles GET /api/v1/media/:id
// @Summary Get media record by ID
// @Tags media
// @Produce json
// @Security BearerAuth
// @Param id path int true "Media ID"
// @Success 200 {object} response.MediaResponse
// @Router /api/v1/media/{id} [get]
func (h *MediaHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	m, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Media not found")
	}

	return c.JSON(http.StatusOK, mappers.ToMediaResponse(m))
}

// FindAll handles GET /api/v1/media
// @Summary List media records
// @Tags media
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.MediaResponse
// @Router /api/v1/media [get]
func (h *MediaHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	mediaFiles, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToMediaResponseList(mediaFiles))
}

// Delete handles DELETE /api/v1/media/:id
// @Summary Delete media record
// @Tags media
// @Security BearerAuth
// @Param id path int true "Media ID"
// @Success 204 "No Content"
// @Router /api/v1/media/{id} [delete]
func (h *MediaHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

