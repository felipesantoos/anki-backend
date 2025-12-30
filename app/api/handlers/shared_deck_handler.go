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

// SharedDeckHandler handles marketplace-related HTTP requests
type SharedDeckHandler struct {
	service primary.ISharedDeckService
}

// NewSharedDeckHandler creates a new SharedDeckHandler instance
func NewSharedDeckHandler(service primary.ISharedDeckService) *SharedDeckHandler {
	return &SharedDeckHandler{
		service: service,
	}
}

// Create handles POST /api/v1/marketplace/decks
// @Summary Publish a deck to marketplace
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateSharedDeckRequest true "Share request"
// @Success 201 {object} response.SharedDeckResponse
// @Router /api/v1/marketplace/decks [post]
func (h *SharedDeckHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateSharedDeckRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	sd, err := h.service.Create(ctx, userID, req.Name, req.Description, req.Category, req.PackagePath, req.PackageSize, req.Tags)
	if err != nil {
		c.Logger().Errorf("Create shared deck error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToSharedDeckResponse(sd))
}

// FindAll handles GET /api/v1/marketplace/decks
// @Summary List public shared decks
// @Tags marketplace
// @Produce json
// @Param category query string false "Category filter"
// @Param tags query []string false "Tags filter"
// @Success 200 {array} response.SharedDeckResponse
// @Router /api/v1/marketplace/decks [get]
func (h *SharedDeckHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	category := c.QueryParam("category")
	var catPtr *string
	if category != "" {
		catPtr = &category
	}
	tags := c.QueryParams()["tags"]

	decks, err := h.service.FindAll(ctx, catPtr, tags)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToSharedDeckResponseList(decks))
}

// FindByID handles GET /api/v1/marketplace/decks/:id
// @Summary Get shared deck by ID
// @Tags marketplace
// @Produce json
// @Param id path int true "Shared Deck ID"
// @Success 200 {object} response.SharedDeckResponse
// @Router /api/v1/marketplace/decks/{id} [get]
func (h *SharedDeckHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c) // Optional, 0 if not auth
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	sd, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Shared deck not found")
	}

	return c.JSON(http.StatusOK, mappers.ToSharedDeckResponse(sd))
}

// Update handles PUT /api/v1/marketplace/decks/:id
// @Summary Update shared deck info
// @Tags marketplace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shared Deck ID"
// @Param request body request.UpdateSharedDeckRequest true "Update request"
// @Success 200 {object} response.SharedDeckResponse
// @Router /api/v1/marketplace/decks/{id} [put]
func (h *SharedDeckHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateSharedDeckRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	sd, err := h.service.Update(ctx, userID, id, req.Name, req.Description, req.Category, req.IsPublic, req.Tags)
	if err != nil {
		c.Logger().Errorf("Update shared deck error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToSharedDeckResponse(sd))
}

// Delete handles DELETE /api/v1/marketplace/decks/:id
// @Summary Remove deck from marketplace
// @Tags marketplace
// @Security BearerAuth
// @Param id path int true "Shared Deck ID"
// @Success 204 "No Content"
// @Router /api/v1/marketplace/decks/{id} [delete]
func (h *SharedDeckHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Download handles POST /api/v1/marketplace/decks/:id/download
// @Summary Record a deck download
// @Tags marketplace
// @Security BearerAuth
// @Param id path int true "Shared Deck ID"
// @Success 204 "No Content"
// @Router /api/v1/marketplace/decks/{id}/download [post]
func (h *SharedDeckHandler) Download(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.IncrementDownloadCount(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

