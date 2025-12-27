package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

// ToDomain converts a RegisterRequest DTO to domain entities
// Note: Password validation and hashing is handled in the domain layer (valueobjects.Password)
func ToDomain(req *request.RegisterRequest) (string, string) {
	return req.Email, req.Password
}

// ToRegisterResponse converts a User entity to RegisterResponse DTO
func ToRegisterResponse(user *entities.User) *response.RegisterResponse {
	return &response.RegisterResponse{
		User: response.UserData{
			ID:            user.ID,
			Email:         user.Email.Value(),
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
		},
	}
}
