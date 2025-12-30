package response

import "time"

// UserResponse represents the full user account information
type UserResponse struct {
	ID            int64     `json:"id" example:"1"`
	Email         string    `json:"email" example:"user@example.com"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

