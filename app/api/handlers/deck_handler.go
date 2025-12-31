package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// DeckHandler handles deck-related HTTP requests
type DeckHandler struct {
	deckService primary.IDeckService
}

// NewDeckHandler creates a new DeckHandler instance
func NewDeckHandler(deckService primary.IDeckService) *DeckHandler {
	return &DeckHandler{
		deckService: deckService,
	}
}

// Create handles POST /api/v1/decks
// @Summary Create a new deck
// @Description Creates a new deck for the authenticated user
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateDeckRequest true "Deck creation request"
// @Success 201 {object} response.DeckResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/decks [post]
func (h *DeckHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateDeckRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	d, err := h.deckService.Create(ctx, userID, req.Name, req.ParentID, req.OptionsJSON)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, mappers.ToDeckResponse(d))
}

// FindByID handles GET /api/v1/decks/:id
// @Summary Get deck by ID
// @Description Returns a deck by its ID for the authenticated user
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deck ID"
// @Success 200 {object} response.DeckResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/decks/{id} [get]
func (h *DeckHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	d, err := h.deckService.FindByID(ctx, userID, id)
	if err != nil {
		return handleDeckError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToDeckResponse(d))
}

// FindAll handles GET /api/v1/decks
// @Summary List all decks
// @Description Returns all decks belonging to the authenticated user
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.DeckResponse
// @Router /api/v1/decks [get]
func (h *DeckHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	decks, err := h.deckService.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, mappers.ToDeckResponseList(decks))
}

// Update handles PUT /api/v1/decks/:id
// @Summary Update deck
// @Description Updates an existing deck's name, parent or options
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deck ID"
// @Param request body request.UpdateDeckRequest true "Deck update request"
// @Success 200 {object} response.DeckResponse
// @Router /api/v1/decks/{id} [put]
func (h *DeckHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateDeckRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	d, err := h.deckService.Update(ctx, userID, id, req.Name, req.ParentID, req.OptionsJSON)
	if err != nil {
		return handleDeckError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToDeckResponse(d))
}

// GetOptions handles GET /api/v1/decks/:id/options
// @Summary Get deck options
// @Description Returns the configuration options for a specific deck
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deck ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/decks/{id}/options [get]
func (h *DeckHandler) GetOptions(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	d, err := h.deckService.FindByID(ctx, userID, id)
	if err != nil {
		return handleDeckError(err)
	}

	var options map[string]interface{}
	if err := json.Unmarshal([]byte(d.GetOptionsJSON()), &options); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse deck options")
	}

	return c.JSON(http.StatusOK, options)
}

// UpdateOptions handles PUT /api/v1/decks/:id/options
// @Summary Update deck options
// @Description Updates the configuration options for a specific deck
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deck ID"
// @Param request body map[string]interface{} true "Deck options update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/decks/{id}/options [put]
func (h *DeckHandler) UpdateOptions(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var options map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&options); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to marshal deck options")
	}

	d, err := h.deckService.UpdateOptions(ctx, userID, id, string(optionsJSON))
	if err != nil {
		return handleDeckError(err)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(d.GetOptionsJSON()), &result)

	return c.JSON(http.StatusOK, result)
}

// handleDeckError maps service-level deck errors to HTTP errors
func handleDeckError(err error) error {
	if errors.Is(err, deck.ErrDeckNotFound) || errors.Is(err, ownership.ErrResourceNotFound) || errors.Is(err, ownership.ErrAccessDenied) {
		return echo.NewHTTPError(http.StatusNotFound, "Deck not found")
	}
	if errors.Is(err, deck.ErrCircularDependency) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// For other errors, let the custom error handler deal with it or default to 500
	return err
}

// Delete handles DELETE /api/v1/decks/:id
// @Summary Delete deck
// @Description Soft deletes a deck and its sub-decks
// @Tags decks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deck ID"
// @Success 204 "No Content"
// @Router /api/v1/decks/{id} [delete]
func (h *DeckHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.deckService.Delete(ctx, userID, id); err != nil {
		return handleDeckError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

