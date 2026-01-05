package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
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

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
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

// handleRegisterError handles errors from the auth service and converts them to appropriate HTTP errors
func handleRegisterError(err error) *echo.HTTPError {
	if errors.Is(err, user.ErrEmailAlreadyExists) {
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

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	// Extract IP address and user agent from request
	ipAddress := c.RealIP()
	if ipAddress == "" {
		ipAddress = c.Request().RemoteAddr
	}
	userAgent := c.Request().UserAgent()

	// Call service
	resp, err := h.authService.Login(ctx, req.Email, req.Password, ipAddress, userAgent)
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

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
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

// VerifyEmail handles GET /api/v1/auth/verify-email requests
// @Summary Verify user email
// @Description Verifies a user's email address using a verification token sent via email
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Email verification token"
// @Success 200 {object} map[string]string "Email verified successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request - token is required"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract token from query parameter
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Token is required")
	}

	// Call service
	err := h.authService.VerifyEmail(ctx, token)
	if err != nil {
		return handleVerifyEmailError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

// ResendVerificationEmail handles POST /api/v1/auth/resend-verification requests
// @Summary Resend verification email
// @Description Resends the email verification email to the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ResendVerificationRequest true "Resend verification request"
// @Success 200 {object} map[string]string "Verification email sent successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 409 {object} response.ErrorResponse "Email already verified"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/resend-verification [post]
func (h *AuthHandler) ResendVerificationEmail(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request body to DTO
	var req request.ResendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request using validator middleware
	if err := c.Validate(&req); err != nil {
		return err // Returns HTTP 400 with validation error message
	}

	// Call service
	err := h.authService.ResendVerificationEmail(ctx, req.Email)
	if err != nil {
		return handleResendVerificationError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Verification email sent successfully",
	})
}

func handleVerifyEmailError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrInvalidToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
	}
	if errors.Is(err, authService.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify email")
}

func handleResendVerificationError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	if strings.Contains(err.Error(), "already verified") {
		return echo.NewHTTPError(http.StatusConflict, "Email already verified")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to resend verification email")
}

// RequestPasswordReset handles POST /api/v1/auth/request-password-reset requests
// @Summary Request password reset
// @Description Sends a password reset email to the user. Always returns success to avoid revealing if email exists.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RequestPasswordResetRequest true "Password reset request"
// @Success 200 {object} map[string]string "Password reset email sent successfully (if email exists)"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/request-password-reset [post]
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request body to DTO
	var req request.RequestPasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := validateRequestPasswordResetRequest(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Call service - always returns success to avoid revealing email existence
	err := h.authService.RequestPasswordReset(ctx, req.Email)
	if err != nil {
		// Even if there's an error, return success to avoid revealing information
		// In production, this should be logged
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password requests
// @Summary Reset password
// @Description Resets user password using a reset token received via email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} map[string]string "Password reset successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request (e.g., invalid password, password mismatch)"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired token"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request body to DTO
	var req request.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := validateResetPasswordRequest(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Call service
	err := h.authService.ResetPassword(ctx, req.Token, req.NewPassword)
	if err != nil {
		return handleResetPasswordError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password reset successfully. Please log in with your new password.",
	})
}

func validateRequestPasswordResetRequest(req *request.RequestPasswordResetRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

func validateResetPasswordRequest(req *request.ResetPasswordRequest) error {
	var errors []string

	if strings.TrimSpace(req.Token) == "" {
		errors = append(errors, "token is required")
	}

	if strings.TrimSpace(req.NewPassword) == "" {
		errors = append(errors, "new_password is required")
	}

	if len(req.NewPassword) < 8 {
		errors = append(errors, "password must have at least 8 characters")
	}

	if req.NewPassword != req.PasswordConfirm {
		errors = append(errors, "password confirmation does not match")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func handleResetPasswordError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrInvalidToken) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
	}
	if errors.Is(err, authService.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	if errors.Is(err, authService.ErrInvalidPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
}

// ChangePassword handles POST /api/v1/auth/change-password requests
// @Summary Change password
// @Description Changes user password. Requires authentication via JWT Bearer token.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer JWT token (access token)"
// @Param request body request.ChangePasswordRequest true "Change password request"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request (e.g., invalid password, password mismatch)"
// @Failure 401 {object} response.ErrorResponse "Not authenticated or current password is incorrect"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract user ID from context (set by auth middleware)
	userID := middlewares.GetUserID(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Bind request body to DTO
	var req request.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := validateChangePasswordRequest(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Call service
	err := h.authService.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return handleChangePasswordError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully. Please log in with your new password.",
	})
}

func validateChangePasswordRequest(req *request.ChangePasswordRequest) error {
	var errors []string

	if strings.TrimSpace(req.CurrentPassword) == "" {
		errors = append(errors, "current_password is required")
	}

	if strings.TrimSpace(req.NewPassword) == "" {
		errors = append(errors, "new_password is required")
	}

	if len(req.NewPassword) < 8 {
		errors = append(errors, "password must have at least 8 characters")
	}

	if req.NewPassword != req.PasswordConfirm {
		errors = append(errors, "password confirmation does not match")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

func handleChangePasswordError(err error) *echo.HTTPError {
	if errors.Is(err, authService.ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Current password is incorrect")
	}
	if errors.Is(err, authService.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	if errors.Is(err, authService.ErrInvalidPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to change password")
}
