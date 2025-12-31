package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
)

// ToDomain converts a RegisterRequest DTO to domain entities
// Note: Password validation and hashing is handled in the domain layer (valueobjects.Password)
func ToDomain(req *request.RegisterRequest) (string, string) {
	return req.Email, req.Password
}

// ToRegisterResponse converts a User entity to RegisterResponse DTO
func ToRegisterResponse(user *user.User) *response.RegisterResponse {
	return &response.RegisterResponse{
		User: response.UserData{
			ID:            user.GetID(),
			Email:         user.GetEmail().Value(),
			EmailVerified: user.GetEmailVerified(),
			CreatedAt:     user.GetCreatedAt(),
			LastLoginAt:   user.GetLastLoginAt(),
		},
	}
}

// ToLoginResponse converts a User entity and tokens to LoginResponse DTO
func ToLoginResponse(user *user.User, accessToken, refreshToken string, expiresIn int) *response.LoginResponse {
	return &response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User: response.UserData{
			ID:            user.GetID(),
			Email:         user.GetEmail().Value(),
			EmailVerified: user.GetEmailVerified(),
			CreatedAt:     user.GetCreatedAt(),
			LastLoginAt:   user.GetLastLoginAt(),
		},
	}
}

// ToTokenResponse converts an access token and expiry to TokenResponse DTO
func ToTokenResponse(accessToken string, expiresIn int) *response.TokenResponse {
	return &response.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}
}
