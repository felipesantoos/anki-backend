package request

// UpdateUserRequest represents the request payload to update user account info
type UpdateUserRequest struct {
	// New email address
	Email string `json:"email" example:"novo@email.com" validate:"required,email"`
}

