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

// NoteTypeHandler handles note type-related HTTP requests
type NoteTypeHandler struct {
	service primary.INoteTypeService
}

// NewNoteTypeHandler creates a new NoteTypeHandler instance
func NewNoteTypeHandler(service primary.INoteTypeService) *NoteTypeHandler {
	return &NoteTypeHandler{
		service: service,
	}
}

// Create handles POST /api/v1/note-types
// @Summary Create a note type
// @Tags note-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateNoteTypeRequest true "Note type creation request"
// @Success 201 {object} response.NoteTypeResponse
// @Router /api/v1/note-types [post]
func (h *NoteTypeHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateNoteTypeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	nt, err := h.service.Create(ctx, userID, req.Name, req.FieldsJSON, req.CardTypesJSON, req.TemplatesJSON)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToNoteTypeResponse(nt))
}

// FindByID handles GET /api/v1/note-types/:id
// @Summary Get note type by ID
// @Tags note-types
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note Type ID"
// @Success 200 {object} response.NoteTypeResponse
// @Router /api/v1/note-types/{id} [get]
func (h *NoteTypeHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	nt, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Note type not found")
	}

	return c.JSON(http.StatusOK, mappers.ToNoteTypeResponse(nt))
}

// FindAll handles GET /api/v1/note-types
// @Summary List note types
// @Description Returns all note types belonging to the authenticated user, with optional search filter by name (case-insensitive, partial match).
// @Tags note-types
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search note types by name (case-insensitive, partial match)"
// @Success 200 {array} response.NoteTypeResponse
// @Router /api/v1/note-types [get]
func (h *NoteTypeHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.ListNoteTypesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	noteTypes, err := h.service.FindByUserID(ctx, userID, req.Search)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToNoteTypeResponseList(noteTypes))
}

// Update handles PUT /api/v1/note-types/:id
// @Summary Update note type
// @Tags note-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note Type ID"
// @Param request body request.UpdateNoteTypeRequest true "Update request"
// @Success 200 {object} response.NoteTypeResponse
// @Router /api/v1/note-types/{id} [put]
func (h *NoteTypeHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateNoteTypeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	nt, err := h.service.Update(ctx, userID, id, req.Name, req.FieldsJSON, req.CardTypesJSON, req.TemplatesJSON)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToNoteTypeResponse(nt))
}

// Delete handles DELETE /api/v1/note-types/:id
// @Summary Delete note type
// @Tags note-types
// @Security BearerAuth
// @Param id path int true "Note Type ID"
// @Success 204 "No Content"
// @Router /api/v1/note-types/{id} [delete]
func (h *NoteTypeHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

