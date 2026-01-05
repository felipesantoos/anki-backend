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

// DeckOptionsPresetHandler handles deck options preset-related HTTP requests
type DeckOptionsPresetHandler struct {
	service primary.IDeckOptionsPresetService
}

// NewDeckOptionsPresetHandler creates a new DeckOptionsPresetHandler instance
func NewDeckOptionsPresetHandler(service primary.IDeckOptionsPresetService) *DeckOptionsPresetHandler {
	return &DeckOptionsPresetHandler{
		service: service,
	}
}

// Create handles POST /api/v1/deck-options-presets
// @Summary Create a new deck options preset
// @Description Creates a new deck options preset for the authenticated user
// @Tags deck-options-presets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateDeckOptionsPresetRequest true "Preset creation request"
// @Success 201 {object} response.DeckOptionsPresetResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/deck-options-presets [post]
func (h *DeckOptionsPresetHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateDeckOptionsPresetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	p, err := h.service.Create(ctx, userID, req.Name, req.OptionsJSON)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, mappers.ToDeckOptionsPresetResponse(p))
}

// FindAll handles GET /api/v1/deck-options-presets
// @Summary List all deck options presets
// @Description Returns all deck options presets belonging to the authenticated user
// @Tags deck-options-presets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.DeckOptionsPresetResponse
// @Router /api/v1/deck-options-presets [get]
func (h *DeckOptionsPresetHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	presets, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, mappers.ToDeckOptionsPresetResponseList(presets))
}

// Update handles PUT /api/v1/deck-options-presets/:id
// @Summary Update deck options preset
// @Description Updates an existing deck options preset's name or options
// @Tags deck-options-presets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Preset ID"
// @Param request body request.UpdateDeckOptionsPresetRequest true "Preset update request"
// @Success 200 {object} response.DeckOptionsPresetResponse
// @Router /api/v1/deck-options-presets/{id} [put]
func (h *DeckOptionsPresetHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateDeckOptionsPresetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	p, err := h.service.Update(ctx, userID, id, req.Name, req.OptionsJSON)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, mappers.ToDeckOptionsPresetResponse(p))
}

// Delete handles DELETE /api/v1/deck-options-presets/:id
// @Summary Delete deck options preset
// @Description Deletes a deck options preset
// @Tags deck-options-presets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Preset ID"
// @Success 204 "No Content"
// @Router /api/v1/deck-options-presets/{id} [delete]
func (h *DeckOptionsPresetHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// ApplyToDecks handles POST /api/v1/deck-options-presets/:id/apply
// @Summary Apply preset to decks
// @Description Applies a preset's options to a list of decks
// @Tags deck-options-presets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Preset ID"
// @Param request body request.ApplyDeckOptionsPresetRequest true "Apply preset request"
// @Success 200 "OK"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/deck-options-presets/{id}/apply [post]
func (h *DeckOptionsPresetHandler) ApplyToDecks(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.ApplyDeckOptionsPresetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	if err := h.service.ApplyToDecks(ctx, userID, id, req.DeckIDs); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

