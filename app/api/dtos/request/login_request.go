package request

// LoginRequest represents the request payload for user login
// @Description Request payload for user login endpoint
type LoginRequest struct {
	// Email address
	Email string `json:"email" validate:"required,email" example:"usuario@example.com"`

	// User password
	Password string `json:"password" validate:"required,min=8" example:"senhaSegura123"`
}

