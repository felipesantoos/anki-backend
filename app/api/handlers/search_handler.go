package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
	service primary.ISearchService
}

// NewSearchHandler creates a new SearchHandler instance
func NewSearchHandler(service primary.ISearchService) *SearchHandler {
	return &SearchHandler{
		service: service,
	}
}

// SearchAdvanced handles POST /api/v1/search/advanced
// @Summary Perform advanced search using Anki syntax
// @Tags search
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.AdvancedSearchRequest true "Advanced search request"
// @Success 200 {object} response.SearchResult
// @Router /api/v1/search/advanced [post]
func (h *SearchHandler) SearchAdvanced(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.AdvancedSearchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query is required")
	}
	if req.Type != "notes" && req.Type != "cards" {
		return echo.NewHTTPError(http.StatusBadRequest, "Type must be 'notes' or 'cards'")
	}

	// Call service
	result, err := h.service.SearchAdvanced(ctx, userID, req.Query, req.Type, req.Limit, req.Offset)
	if err != nil {
		return handleSearchError(err)
	}

	// Map results to response DTOs
	responseData := make([]interface{}, len(result.Data))
	for i, item := range result.Data {
		switch v := item.(type) {
		case *note.Note:
			responseData[i] = mappers.ToNoteResponse(v)
		case *card.Card:
			responseData[i] = mappers.ToCardResponse(v)
		default:
			// Unknown type, return as-is
			responseData[i] = item
		}
	}

	return c.JSON(http.StatusOK, response.SearchResult{
		Data:  responseData,
		Total: result.Total,
	})
}

func handleSearchError(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "not found") {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if strings.Contains(err.Error(), "invalid regex") || strings.Contains(err.Error(), "regex pattern") {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "parse") {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Default to internal server error
	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
}

