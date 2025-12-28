package request

// ChangePasswordRequest represents the request payload for changing password when authenticated
// @Description Request payload for changing password (requires authentication)
type ChangePasswordRequest struct {
	// Current password (required for security)
	// @Example currentPassword123
	CurrentPassword string `json:"current_password" validate:"required" example:"currentPassword123"`

	// New password (minimum 8 characters, must contain at least one letter and one number)
	// @Example newSecurePassword123
	NewPassword string `json:"new_password" validate:"required,min=8" example:"newSecurePassword123"`

	// Password confirmation (must match new_password)
	// @Example newSecurePassword123
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=NewPassword" example:"newSecurePassword123"`
}

