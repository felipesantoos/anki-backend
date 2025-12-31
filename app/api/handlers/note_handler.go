package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// NoteHandler handles note-related HTTP requests
type NoteHandler struct {
	service primary.INoteService
}

// NewNoteHandler creates a new NoteHandler instance
func NewNoteHandler(service primary.INoteService) *NoteHandler {
	return &NoteHandler{
		service: service,
	}
}

// Create handles POST /api/v1/notes
// @Summary Create a note
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateNoteRequest true "Note creation request"
// @Success 201 {object} response.NoteResponse
// @Router /api/v1/notes [post]
func (h *NoteHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateNoteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	n, err := h.service.Create(ctx, userID, req.NoteTypeID, req.DeckID, req.FieldsJSON, req.Tags)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusCreated, mappers.ToNoteResponse(n))
}

// FindByID handles GET /api/v1/notes/:id
// @Summary Get note by ID
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Success 200 {object} response.NoteResponse
// @Router /api/v1/notes/{id} [get]
func (h *NoteHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	n, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToNoteResponse(n))
}

// FindAll handles GET /api/v1/notes
// @Summary List notes
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.NoteResponse
// @Router /api/v1/notes [get]
func (h *NoteHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	notes, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToNoteResponseList(notes))
}

// Update handles PUT /api/v1/notes/:id
// @Summary Update note
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param request body request.UpdateNoteRequest true "Update request"
// @Success 200 {object} response.NoteResponse
// @Router /api/v1/notes/{id} [put]
func (h *NoteHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateNoteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	n, err := h.service.Update(ctx, userID, id, req.FieldsJSON, req.Tags)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToNoteResponse(n))
}

// Delete handles DELETE /api/v1/notes/:id
// @Summary Delete note
// @Tags notes
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Success 204 "No Content"
// @Router /api/v1/notes/{id} [delete]
func (h *NoteHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return handleNoteError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// AddTag handles POST /api/v1/notes/:id/tags
// @Summary Add tag to note
// @Tags notes
// @Accept json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param request body request.AddTagRequest true "Tag request"
// @Success 204 "No Content"
// @Router /api/v1/notes/{id}/tags [post]
func (h *NoteHandler) AddTag(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.AddTagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.service.AddTag(ctx, userID, id, req.Tag); err != nil {
		return handleNoteError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RemoveTag handles DELETE /api/v1/notes/:id/tags/:tag
// @Summary Remove tag from note
// @Tags notes
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param tag path string true "Tag name"
// @Success 204 "No Content"
// @Router /api/v1/notes/{id}/tags/{tag} [delete]
func (h *NoteHandler) RemoveTag(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tag := c.Param("tag")

	if err := h.service.RemoveTag(ctx, userID, id, tag); err != nil {
		return handleNoteError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func handleNoteError(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "not found") {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	// Default to bad request for other validation errors
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

