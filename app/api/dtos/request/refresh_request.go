package request

// RefreshRequest represents the request payload for refreshing an access token
// @Description Request payload for refresh token endpoint
type RefreshRequest struct {
	// Refresh token
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

