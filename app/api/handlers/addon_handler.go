package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// AddOnHandler handles add-on related HTTP requests
type AddOnHandler struct {
	service primary.IAddOnService
}

// NewAddOnHandler creates a new AddOnHandler instance
func NewAddOnHandler(service primary.IAddOnService) *AddOnHandler {
	return &AddOnHandler{
		service: service,
	}
}

// Install handles POST /api/v1/addons
// @Summary Install an add-on
// @Tags addons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.InstallAddOnRequest true "Install request"
// @Success 201 {object} response.AddOnResponse
// @Router /api/v1/addons [post]
func (h *AddOnHandler) Install(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.InstallAddOnRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	a, err := h.service.Install(ctx, userID, req.Code, req.Name, req.Version, req.ConfigJSON)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToAddOnResponse(a))
}

// FindAll handles GET /api/v1/addons
// @Summary List installed add-ons
// @Tags addons
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.AddOnResponse
// @Router /api/v1/addons [get]
func (h *AddOnHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	addOns, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToAddOnResponseList(addOns))
}

// UpdateConfig handles PUT /api/v1/addons/:code/config
// @Summary Update add-on configuration
// @Tags addons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Add-on code"
// @Param request body request.UpdateAddOnConfigRequest true "Config request"
// @Success 200 {object} response.AddOnResponse
// @Router /api/v1/addons/{code}/config [put]
func (h *AddOnHandler) UpdateConfig(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	code := c.Param("code")

	var req request.UpdateAddOnConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	a, err := h.service.UpdateConfig(ctx, userID, code, req.ConfigJSON)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToAddOnResponse(a))
}

// Toggle handles POST /api/v1/addons/:code/toggle
// @Summary Enable or disable an add-on
// @Tags addons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Add-on code"
// @Param request body request.ToggleAddOnRequest true "Toggle request"
// @Success 200 {object} response.AddOnResponse
// @Router /api/v1/addons/{code}/toggle [post]
func (h *AddOnHandler) Toggle(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	code := c.Param("code")

	var req request.ToggleAddOnRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	a, err := h.service.ToggleEnabled(ctx, userID, code, req.Enabled)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToAddOnResponse(a))
}

// Uninstall handles DELETE /api/v1/addons/:code
// @Summary Uninstall an add-on
// @Tags addons
// @Security BearerAuth
// @Param code path string true "Add-on code"
// @Success 204 "No Content"
// @Router /api/v1/addons/{code} [delete]
func (h *AddOnHandler) Uninstall(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	code := c.Param("code")

	if err := h.service.Uninstall(ctx, userID, code); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

