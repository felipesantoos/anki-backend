package response

import "time"

// RegisterResponse represents the response payload after successful user registration
// @Description Response payload for user registration endpoint
type RegisterResponse struct {
	// User information
	User UserData `json:"user"`
}

// UserData contains the user information returned after registration
type UserData struct {
	// User ID
	ID int64 `json:"id" example:"1"`

	// Email address
	Email string `json:"email" example:"usuario@example.com"`

	// Whether the email has been verified
	EmailVerified bool `json:"email_verified" example:"false"`

	// Timestamp when the user was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp of the last login
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2024-01-15T10:30:00Z"`
}
