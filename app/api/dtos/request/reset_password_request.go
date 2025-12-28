package request

// ResetPasswordRequest represents the request payload for resetting a password
// @Description Request payload for resetting password using a reset token
type ResetPasswordRequest struct {
	// Password reset token received via email
	// @Example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// New password (minimum 8 characters, must contain at least one letter and one number)
	// @Example newSecurePassword123
	NewPassword string `json:"new_password" validate:"required,min=8" example:"newSecurePassword123"`

	// Password confirmation (must match new_password)
	// @Example newSecurePassword123
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=NewPassword" example:"newSecurePassword123"`
}

