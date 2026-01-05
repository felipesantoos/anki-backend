package request

// ResendVerificationRequest represents the request payload for resending verification email
// @Description Request payload for resending email verification
type ResendVerificationRequest struct {
	// Email address to resend verification email to
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

