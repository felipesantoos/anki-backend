package request

// RequestPasswordResetRequest represents the request payload for requesting a password reset
// @Description Request payload for requesting a password reset email
type RequestPasswordResetRequest struct {
	// Email address to send password reset link to
	// @Example user@example.com
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

