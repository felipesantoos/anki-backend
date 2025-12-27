package handlers

import (
	"errors"
	"fmt"
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

// Login handles POST /api/v1/auth/login requests
// @Summary Login user
// @Description Authenticates a user with email and password and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "Login request"
// @Success 200 {object} response.LoginResponse "Login successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Invalid credentials"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request body to DTO
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := validateLoginRequest(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Call service
	resp, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return handleLoginError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

// RefreshToken handles POST /api/v1/auth/refresh requests
// @Summary Refresh access token
// @Description Generates a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RefreshRequest true "Refresh token request"
// @Success 200 {object} response.TokenResponse "Token refreshed successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request body to DTO
	var req request.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}

	// Call service
	resp, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return handleRefreshError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout requests
// @Summary Logout user
// @Description Invalidates both access token and refresh token. Access token should be provided in Authorization header as Bearer token. Refresh token is optional and can be provided in the request body.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token (access token)"
// @Param request body request.RefreshRequest false "Logout request (refresh_token optional if access token is provided in header)"
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request - either Authorization header with Bearer token or refresh_token in body is required"
// @Failure 401 {object} response.ErrorResponse "Invalid token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract access token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	var accessToken string
	if authHeader != "" {
		// Extract token from "Bearer <token>" format
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			accessToken = authHeader[7:]
		}
	}

	// Bind request body to DTO (for refresh token, optional)
	var req request.RefreshRequest
	// Don't fail if body is empty - access token from header is enough
	_ = c.Bind(&req)

	// At least one token must be provided
	if accessToken == "" && req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Either Authorization header with Bearer token or refresh_token in body is required")
	}

	// Call service
	err := h.authService.Logout(ctx, accessToken, req.RefreshToken)
	if err != nil {
		return handleLogoutError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// validateLoginRequest validates the login request
func validateLoginRequest(req *request.LoginRequest) error {
	var errors []string

	if req.Email == "" {
		errors = append(errors, "email is required")
	}

	if req.Password == "" {
		errors = append(errors, "password is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, ", "))
	}

	return nil
}

// handleLoginError handles errors from the login service and converts them to appropriate HTTP errors
func handleLoginError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	if errors.Is(err, authService.ErrInvalidEmail) || errors.Is(err, authService.ErrInvalidPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// For other errors, return 500
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to login")
}

// handleRefreshError handles errors from the refresh token service and converts them to appropriate HTTP errors
func handleRefreshError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrInvalidToken) || errors.Is(err, authService.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
	}

	// For other errors, return 500
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh token")
}

// handleLogoutError handles errors from the logout service and converts them to appropriate HTTP errors
func handleLogoutError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrInvalidToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	// For other errors, return 500
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout")
}
