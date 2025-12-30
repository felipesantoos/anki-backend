package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	userSvc "github.com/felipesantos/anki-backend/core/services/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := userSvc.NewUserService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		oldEmail, _ := valueobjects.NewEmail("old@example.com")
		u, _ := user.NewBuilder().WithID(userID).WithEmail(oldEmail).Build()
		newEmailStr := "new@example.com"

		mockRepo.On("FindByID", ctx, userID).Return(u, nil).Once()
		mockRepo.On("ExistsByEmail", ctx, newEmailStr).Return(false, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, newEmailStr)

		assert.NoError(t, err)
		assert.Equal(t, newEmailStr, result.GetEmail().Value())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Email Already Taken", func(t *testing.T) {
		oldEmail, _ := valueobjects.NewEmail("old@example.com")
		u, _ := user.NewBuilder().WithID(userID).WithEmail(oldEmail).Build()
		newEmailStr := "taken@example.com"

		mockRepo.On("FindByID", ctx, userID).Return(u, nil).Once()
		mockRepo.On("ExistsByEmail", ctx, newEmailStr).Return(true, nil).Once()

		result, err := service.Update(ctx, userID, newEmailStr)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already in use")
		mockRepo.AssertExpectations(t)
	})
}

