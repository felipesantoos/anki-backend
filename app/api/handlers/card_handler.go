package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/pkg/ownership"
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

// FindAll handles GET /api/v1/cards
// @Summary List cards
// @Description List cards with optional filters and pagination
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param deck_id query int false "Filter by deck ID"
// @Param state query string false "Filter by state (new, learn, review, relearn)"
// @Param flag query int false "Filter by flag (0-7)"
// @Param suspended query bool false "Filter by suspended"
// @Param buried query bool false "Filter by buried"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} response.ListCardsResponse
// @Router /api/v1/cards [get]
func (h *CardHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.ListCardsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	// Convert ListCardsRequest to CardFilters
	filters := card.CardFilters{
		DeckID:    req.DeckID,
		State:     req.State,
		Flag:      req.Flag,
		Suspended: req.Suspended,
		Buried:    req.Buried,
	}

	// Apply pagination defaults and calculate offset
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	filters.Limit = limit
	filters.Offset = offset

	// Call service
	cards, total, err := h.service.FindAll(ctx, userID, filters)
	if err != nil {
		return err
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	// Build response
	return c.JSON(http.StatusOK, response.ListCardsResponse{
		Data: mappers.ToCardResponseList(cards),
		Pagination: response.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// FindLeeches handles GET /api/v1/cards/leeches
// @Summary List leech cards
// @Description List cards that are difficult to memorize (leeches)
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} response.ListCardsResponse
// @Router /api/v1/cards/leeches [get]
func (h *CardHandler) FindLeeches(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	cards, total, err := h.service.FindLeeches(ctx, userID, limit, offset)
	if err != nil {
		return handleCardError(err)
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return c.JSON(http.StatusOK, response.ListCardsResponse{
		Data: mappers.ToCardResponseList(cards),
		Pagination: response.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// FindByID handles GET /api/v1/cards/:id
// @Summary Get card by ID
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} response.CardResponse
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id} [get]
func (h *CardHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	card, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return handleCardError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToCardResponse(card))
}

// GetInfo handles GET /api/v1/cards/:id/info
// @Summary Get detailed card information
// @Description Returns detailed card information including note data, deck/note type names, and review history
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} response.CardInfoResponse
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/info [get]
func (h *CardHandler) GetInfo(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	info, err := h.service.GetInfo(ctx, userID, id)
	if err != nil {
		return handleCardError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToCardInfoResponse(info))
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
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/suspend [post]
func (h *CardHandler) Suspend(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	if err := h.service.Suspend(ctx, userID, id); err != nil {
		return handleCardError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Unsuspend handles POST /api/v1/cards/:id/unsuspend
// @Summary Unsuspend card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/unsuspend [post]
func (h *CardHandler) Unsuspend(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	if err := h.service.Unsuspend(ctx, userID, id); err != nil {
		return handleCardError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Bury handles POST /api/v1/cards/:id/bury
// @Summary Bury card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/bury [post]
func (h *CardHandler) Bury(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	if err := h.service.Bury(ctx, userID, id); err != nil {
		return handleCardError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Unbury handles POST /api/v1/cards/:id/unbury
// @Summary Unbury card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/unbury [post]
func (h *CardHandler) Unbury(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	if err := h.service.Unbury(ctx, userID, id); err != nil {
		return handleCardError(err)
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
// @Failure 400 {object} response.ErrorResponse "Invalid card ID or flag value"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/flag [post]
func (h *CardHandler) SetFlag(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	var req request.SetCardFlagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	if err := h.service.SetFlag(ctx, userID, id, req.Flag); err != nil {
		return handleCardError(err)
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

// Reset handles POST /api/v1/cards/:id/reset
// @Summary Reset card
// @Description Resets a card to the new state or completely forgets its history
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Param request body request.ResetCardRequest true "Reset request"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid card ID or reset type"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/reset [post]
func (h *CardHandler) Reset(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	var req request.ResetCardRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.Reset(ctx, userID, id, req.Type); err != nil {
		return handleCardError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// SetDueDate handles POST /api/v1/cards/:id/due
// @Summary Set card due date
// @Description Manually sets the next review date for a card
// @Tags cards
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Param request body request.SetCardDueDateRequest true "Due date request"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid card ID or due date"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/due [post]
func (h *CardHandler) SetDueDate(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	var req request.SetCardDueDateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.SetDueDate(ctx, userID, id, req.Due); err != nil {
		return handleCardError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Reposition handles POST /api/v1/cards/reposition
// @Summary Reposition cards
// @Description Changes the order new cards will appear in
// @Tags cards
// @Security BearerAuth
// @Param request body request.RepositionCardsRequest true "Reposition request"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or parameters"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/reposition [post]
func (h *CardHandler) Reposition(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.RepositionCardsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.Reposition(ctx, userID, req.CardIDs, req.Start, req.Step, req.Shift); err != nil {
		return handleCardError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GetPosition handles GET /api/v1/cards/:id/position
// @Summary Get card position
// @Description Returns the ordinal position of a card
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} response.CardPositionResponse
// @Failure 400 {object} response.ErrorResponse "Invalid card ID"
// @Failure 404 {object} response.ErrorResponse "Card not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/cards/{id}/position [get]
func (h *CardHandler) GetPosition(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse and validate ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid card ID format")
	}

	// Validate that ID is positive
	if id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Card ID must be greater than 0")
	}

	position, err := h.service.GetPosition(ctx, userID, id)
	if err != nil {
		return handleCardError(err)
	}

	return c.JSON(http.StatusOK, response.CardPositionResponse{
		Position: position,
	})
}

// handleCardError maps service-level card errors to HTTP errors
func handleCardError(err error) error {
	if err == nil {
		return nil
	}

	// Check for invalid flag error
	if errors.Is(err, card.ErrInvalidFlag) {
		return echo.NewHTTPError(http.StatusBadRequest, "Flag must be between 0 and 7")
	}

	// Check for resource not found (card doesn't exist or user doesn't have access)
	if errors.Is(err, ownership.ErrResourceNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "Card not found")
	}

	// Check for access denied
	if errors.Is(err, ownership.ErrAccessDenied) {
		return echo.NewHTTPError(http.StatusNotFound, "Card not found")
	}

	// If error message contains "not found", return 404
	if errors.Is(err, errors.New("card not found")) {
		return echo.NewHTTPError(http.StatusNotFound, "Card not found")
	}

	// For other errors, return as-is (may be HTTPError already or will be handled by error middleware)
	return err
}
