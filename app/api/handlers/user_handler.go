package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// UserHandler handles user account-related HTTP requests
type UserHandler struct {
	service primary.IUserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(service primary.IUserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// GetMe handles GET /api/v1/user/me
// @Summary Get current user information
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserResponse
// @Router /api/v1/user/me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	u, err := h.service.FindByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, mappers.ToUserResponse(u))
}

// Update handles PUT /api/v1/user/me
// @Summary Update user account information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateUserRequest true "Update request"
// @Success 200 {object} response.UserResponse
// @Router /api/v1/user/me [put]
func (h *UserHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	u, err := h.service.Update(ctx, userID, req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToUserResponse(u))
}

// Delete handles DELETE /api/v1/user/me
// @Summary Delete user account
// @Description Soft deletes the user account. Irreversible from the API.
// @Tags user
// @Security BearerAuth
// @Success 204 "No Content"
// @Router /api/v1/user/me [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	if err := h.service.Delete(ctx, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

