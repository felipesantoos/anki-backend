package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService primary.IAuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService primary.IAuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles POST /api/v1/auth/register requests
// @Summary Register a new user
// @Description Registers a new user account with email and password. Creates a default deck for the user.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "Registration request"
// @Success 201 {object} response.RegisterResponse "User registered successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 409 {object} response.ErrorResponse "Email already registered"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request body to DTO
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := validateRegisterRequest(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Call service
	user, err := h.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		return handleRegisterError(err)
	}

	// Convert to response
	resp := mappers.ToRegisterResponse(user)

	return c.JSON(http.StatusCreated, resp)
}

// validateRegisterRequest validates the register request
func validateRegisterRequest(req *request.RegisterRequest) error {
	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email is required")
	}

	// Validate password
	if strings.TrimSpace(req.Password) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Password is required")
	}

	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Password must have at least 8 characters")
	}

	// Validate password confirmation
	if req.Password != req.PasswordConfirm {
		return echo.NewHTTPError(http.StatusBadRequest, "Password confirmation does not match")
	}

	return nil
}

// handleRegisterError handles errors from the auth service and converts them to appropriate HTTP errors
func handleRegisterError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrEmailAlreadyExists) {
		return echo.NewHTTPError(http.StatusConflict, "Email already registered")
	}

	if errors.Is(err, authService.ErrInvalidEmail) || errors.Is(err, authService.ErrInvalidPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// For other errors, return 500
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
}
