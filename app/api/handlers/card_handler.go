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

// CardHandler handles card-related HTTP requests
type CardHandler struct {
	service primary.ICardService
}

// NewCardHandler creates a new CardHandler instance
func NewCardHandler(service primary.ICardService) *CardHandler {
	return &CardHandler{
		service: service,
	}
}

// FindByID handles GET /api/v1/cards/:id
// @Summary Get card by ID
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} response.CardResponse
// @Router /api/v1/cards/{id} [get]
func (h *CardHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	card, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, mappers.ToCardResponse(card))
}

// FindByDeckID handles GET /api/v1/decks/:deckID/cards
// @Summary List cards in a deck
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param deckID path int true "Deck ID"
// @Success 200 {array} response.CardResponse
// @Router /api/v1/decks/{deckID}/cards [get]
func (h *CardHandler) FindByDeckID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	deckID, _ := strconv.ParseInt(c.Param("deckID"), 10, 64)

	cards, err := h.service.FindByDeckID(ctx, userID, deckID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, mappers.ToCardResponseList(cards))
}

// FindDueCards handles GET /api/v1/decks/:deckID/cards/due
// @Summary List due cards in a deck
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param deckID path int true "Deck ID"
// @Success 200 {array} response.CardResponse
// @Router /api/v1/decks/{deckID}/cards/due [get]
func (h *CardHandler) FindDueCards(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	deckID, _ := strconv.ParseInt(c.Param("deckID"), 10, 64)

	cards, err := h.service.FindDueCards(ctx, userID, deckID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, mappers.ToCardResponseList(cards))
}

// Suspend handles POST /api/v1/cards/:id/suspend
// @Summary Suspend card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Router /api/v1/cards/{id}/suspend [post]
func (h *CardHandler) Suspend(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Suspend(ctx, userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Unsuspend handles POST /api/v1/cards/:id/unsuspend
// @Summary Unsuspend card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Router /api/v1/cards/{id}/unsuspend [post]
func (h *CardHandler) Unsuspend(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Unsuspend(ctx, userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Bury handles POST /api/v1/cards/:id/bury
// @Summary Bury card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Router /api/v1/cards/{id}/bury [post]
func (h *CardHandler) Bury(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Bury(ctx, userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Unbury handles POST /api/v1/cards/:id/unbury
// @Summary Unbury card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Router /api/v1/cards/{id}/unbury [post]
func (h *CardHandler) Unbury(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Unbury(ctx, userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// SetFlag handles POST /api/v1/cards/:id/flag
// @Summary Set card flag
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Param request body request.SetCardFlagRequest true "Flag request"
// @Success 204 "No Content"
// @Router /api/v1/cards/{id}/flag [post]
func (h *CardHandler) SetFlag(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.SetCardFlagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	if err := h.service.SetFlag(ctx, userID, id, req.Flag); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Delete handles DELETE /api/v1/cards/:id
// @Summary Delete card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Router /api/v1/cards/{id} [delete]
func (h *CardHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

