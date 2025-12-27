package request

// RegisterRequest represents the request payload for user registration
// @Description Request payload for user registration endpoint
type RegisterRequest struct {
	// Email address of the user
	// Required: true
	// Example: "usuario@example.com"
	Email string `json:"email" validate:"required,email" example:"usuario@example.com"`

	// Password for the user account (minimum 8 characters, must contain at least one letter and one number)
	// Required: true
	// Example: "senhaSegura123"
	Password string `json:"password" validate:"required,min=8" example:"senhaSegura123"`

	// Password confirmation (must match password)
	// Required: true
	// Example: "senhaSegura123"
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password" example:"senhaSegura123"`
}
