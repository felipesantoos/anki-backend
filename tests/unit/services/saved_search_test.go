package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	savedsearchsvc "github.com/felipesantos/anki-backend/core/services/savedsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSavedSearchService_Create(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSavedSearchRepository)
	service := savedsearchsvc.NewSavedSearchService(mockRepo)

	userID := int64(1)
	name := "Test Search"
	query := "tags:test"

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindByName", ctx, userID, name).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*savedsearch.SavedSearch")).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("DuplicateName", func(t *testing.T) {
		existing, _ := savedsearch.NewBuilder().
			WithUserID(userID).
			WithName(name).
			Build()
		mockRepo.On("FindByName", ctx, userID, name).Return(existing, nil).Once()

		result, err := service.Create(ctx, userID, name, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})
}

func TestSavedSearchService_FindByUserID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSavedSearchRepository)
	service := savedsearchsvc.NewSavedSearchService(mockRepo)

	userID := int64(1)
	searches := []*savedsearch.SavedSearch{}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return(searches, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, searches, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestSavedSearchService_Update(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSavedSearchRepository)
	service := savedsearchsvc.NewSavedSearchService(mockRepo)

	userID := int64(1)
	id := int64(100)
	name := "New Name"
	query := "new query"

	t.Run("Success", func(t *testing.T) {
		existing, _ := savedsearch.NewBuilder().
			WithUserID(userID).
			WithName("Old Name").
			Build()
		mockRepo.On("FindByID", ctx, userID, id).Return(existing, nil).Once()
		mockRepo.On("FindByName", ctx, userID, name).Return(nil, nil).Once()
		mockRepo.On("Update", ctx, userID, id, existing).Return(nil).Once()

		result, err := service.Update(ctx, userID, id, name, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, id).Return(nil, nil).Once()

		result, err := service.Update(ctx, userID, id, name, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestSavedSearchService_Delete(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSavedSearchRepository)
	service := savedsearchsvc.NewSavedSearchService(mockRepo)

	userID := int64(1)
	id := int64(100)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Delete", ctx, userID, id).Return(nil).Once()

		err := service.Delete(ctx, userID, id)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

