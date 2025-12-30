package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	flagnamesvc "github.com/felipesantos/anki-backend/core/services/flagname"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFlagNameService_FindByUserID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockFlagNameRepository)
	service := flagnamesvc.NewFlagNameService(mockRepo)

	userID := int64(1)
	flag, _ := flagname.NewBuilder().
		WithUserID(userID).
		WithFlagNumber(1).
		WithName("Flag 1").
		Build()
	flags := []*flagname.FlagName{flag}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return(flags, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, flags, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestFlagNameService_Update(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockFlagNameRepository)
	service := flagnamesvc.NewFlagNameService(mockRepo)

	userID := int64(1)
	flagNum := 1
	newName := "New Name"

	t.Run("Success_CreateNew", func(t *testing.T) {
		mockRepo.On("FindByFlagNumber", ctx, userID, flagNum).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*flagname.FlagName")).Return(nil).Once()

		result, err := service.Update(ctx, userID, flagNum, newName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success_UpdateExisting", func(t *testing.T) {
		existing, _ := flagname.NewBuilder().
			WithUserID(userID).
			WithFlagNumber(flagNum).
			WithName("Old Name").
			Build()

		mockRepo.On("FindByFlagNumber", ctx, userID, flagNum).Return(existing, nil).Once()
		mockRepo.On("Save", ctx, userID, existing).Return(nil).Once()

		result, err := service.Update(ctx, userID, flagNum, newName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

