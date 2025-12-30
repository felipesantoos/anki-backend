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

// ProfileHandler handles profile-related HTTP requests
type ProfileHandler struct {
	service primary.IProfileService
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(service primary.IProfileService) *ProfileHandler {
	return &ProfileHandler{
		service: service,
	}
}

// Create handles POST /api/v1/profiles
// @Summary Create a profile
// @Tags profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateProfileRequest true "Profile creation request"
// @Success 201 {object} response.ProfileResponse
// @Router /api/v1/profiles [post]
func (h *ProfileHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	p, err := h.service.Create(ctx, userID, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mappers.ToProfileResponse(p))
}

// FindAll handles GET /api/v1/profiles
// @Summary List profiles
// @Tags profiles
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.ProfileResponse
// @Router /api/v1/profiles [get]
func (h *ProfileHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	profiles, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToProfileResponseList(profiles))
}

// FindByID handles GET /api/v1/profiles/:id
// @Summary Get profile by ID
// @Tags profiles
// @Produce json
// @Security BearerAuth
// @Param id path int true "Profile ID"
// @Success 200 {object} response.ProfileResponse
// @Router /api/v1/profiles/{id} [get]
func (h *ProfileHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	p, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Profile not found")
	}

	return c.JSON(http.StatusOK, mappers.ToProfileResponse(p))
}

// Update handles PUT /api/v1/profiles/:id
// @Summary Update profile
// @Tags profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Profile ID"
// @Param request body request.UpdateProfileRequest true "Update request"
// @Success 200 {object} response.ProfileResponse
// @Router /api/v1/profiles/{id} [put]
func (h *ProfileHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	p, err := h.service.Update(ctx, userID, id, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToProfileResponse(p))
}

// Delete handles DELETE /api/v1/profiles/:id
// @Summary Delete profile
// @Tags profiles
// @Security BearerAuth
// @Param id path int true "Profile ID"
// @Success 204 "No Content"
// @Router /api/v1/profiles/{id} [delete]
func (h *ProfileHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// EnableSync handles POST /api/v1/profiles/:id/sync/enable
// @Summary Enable AnkiWeb sync for profile
// @Tags profiles
// @Accept json
// @Security BearerAuth
// @Param id path int true "Profile ID"
// @Param request body request.EnableSyncRequest true "Sync request"
// @Success 204 "No Content"
// @Router /api/v1/profiles/{id}/sync/enable [post]
func (h *ProfileHandler) EnableSync(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.EnableSyncRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.service.EnableSync(ctx, userID, id, req.Username); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// DisableSync handles POST /api/v1/profiles/:id/sync/disable
// @Summary Disable AnkiWeb sync for profile
// @Tags profiles
// @Security BearerAuth
// @Param id path int true "Profile ID"
// @Success 204 "No Content"
// @Router /api/v1/profiles/{id}/sync/disable [post]
func (h *ProfileHandler) DisableSync(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.DisableSync(ctx, userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

