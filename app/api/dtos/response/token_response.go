package response

// TokenResponse represents the response payload for token refresh
// @Description Response payload for refresh token endpoint
type TokenResponse struct {
	// Access token (JWT)
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Refresh token (JWT) - returned when token rotation is enabled
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Token expiration time in seconds
	ExpiresIn int `json:"expires_in" example:"900"`

	// Token type (usually "Bearer")
	TokenType string `json:"token_type" example:"Bearer"`
}

