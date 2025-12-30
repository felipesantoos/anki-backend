package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
)

// ToUserResponse converts a User domain entity to a UserResponse DTO
func ToUserResponse(u *user.User) *response.UserResponse {
	if u == nil {
		return nil
	}
	return &response.UserResponse{
		ID:            u.GetID(),
		Email:         u.GetEmail().Value(),
		EmailVerified: u.GetEmailVerified(),
		CreatedAt:     u.GetCreatedAt(),
		UpdatedAt:     u.GetUpdatedAt(),
	}
}

