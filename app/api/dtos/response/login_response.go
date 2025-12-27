package response

// LoginResponse represents the response payload after successful login
// @Description Response payload for login endpoint
type LoginResponse struct {
	// Access token (JWT)
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Refresh token
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Token expiration time in seconds
	ExpiresIn int `json:"expires_in" example:"900"`

	// Token type (usually "Bearer")
	TokenType string `json:"token_type" example:"Bearer"`

	// User information
	User UserData `json:"user"`
}

