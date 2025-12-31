package services

import (
	"context"
	"errors"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	deckSvc "github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeckService_Create(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockBackupSvc := new(MockBackupService)
	mockTM := new(MockTransactionManager)
	service := deckSvc.NewDeckService(mockRepo, mockCardRepo, mockBackupSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "New Deck"
		mockRepo.On("Exists", ctx, userID, name, (*int64)(nil)).Return(false, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*deck.Deck")).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, nil, "")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Already Exists", func(t *testing.T) {
		name := "Existing Deck"
		mockRepo.On("Exists", ctx, userID, name, (*int64)(nil)).Return(true, nil).Once()

		result, err := service.Create(ctx, userID, name, nil, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("With Parent Success", func(t *testing.T) {
		name := "Child Deck"
		parentID := int64(10)
		parentDeck, _ := deck.NewBuilder().WithID(parentID).WithUserID(userID).WithName("Parent").Build()

		mockRepo.On("Exists", ctx, userID, name, &parentID).Return(false, nil).Once()
		mockRepo.On("FindByID", ctx, userID, parentID).Return(parentDeck, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, &parentID, "")

		assert.NoError(t, err)
		assert.Equal(t, &parentID, result.GetParentID())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Name Format", func(t *testing.T) {
		_, err := service.Create(ctx, userID, "A::B", nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deck name cannot contain '::'")

		_, err = service.Create(ctx, userID, "::", nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deck name cannot contain '::'")

		_, err = service.Create(ctx, userID, "  ", nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deck name cannot be empty")
	})
}

func TestDeckService_Delete(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockBackupSvc := new(MockBackupService)
	mockTM := new(MockTransactionManager)
	service := deckSvc.NewDeckService(mockRepo, mockCardRepo, mockBackupSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success Delete Cards", func(t *testing.T) {
		deckID := int64(100)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("To Delete").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockBackupSvc.On("CreatePreOperationBackup", ctx, userID).Return(nil, nil).Once()
		mockTM.ExpectTransaction()
		mockCardRepo.On("DeleteByDeckRecursive", mock.Anything, userID, deckID).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, userID, deckID).Return(nil).Once()

		err := service.Delete(ctx, userID, deckID, deck.ActionDeleteCards, nil)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockBackupSvc.AssertExpectations(t)
	})

	t.Run("Backup Failure Fails Deletion", func(t *testing.T) {
		deckID := int64(100)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("To Delete").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockBackupSvc.On("CreatePreOperationBackup", ctx, userID).Return(nil, errors.New("backup failed")).Once()

		err := service.Delete(ctx, userID, deckID, deck.ActionDeleteCards, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create backup before deletion")
		mockBackupSvc.AssertExpectations(t)
	})

	t.Run("Success Move to Default", func(t *testing.T) {
		deckID := int64(100)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("To Delete").Build()
		defaultDeckID := int64(1)
		defaultDeck, _ := deck.NewBuilder().WithID(defaultDeckID).WithUserID(userID).WithName("Default").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockBackupSvc.On("CreatePreOperationBackup", ctx, userID).Return(nil, nil).Once()
		mockRepo.On("FindByUserID", ctx, userID).Return([]*deck.Deck{defaultDeck}, nil).Once()
		mockTM.ExpectTransaction()
		mockCardRepo.On("MoveCards", mock.Anything, userID, deckID, defaultDeckID).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, userID, deckID).Return(nil).Once()

		err := service.Delete(ctx, userID, deckID, deck.ActionMoveToDefault, nil)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockBackupSvc.AssertExpectations(t)
	})

	t.Run("Success Move to Another Deck", func(t *testing.T) {
		deckID := int64(100)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("To Delete").Build()
		targetDeckID := int64(200)
		targetDeck, _ := deck.NewBuilder().WithID(targetDeckID).WithUserID(userID).WithName("Target").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockBackupSvc.On("CreatePreOperationBackup", ctx, userID).Return(nil, nil).Once()
		mockRepo.On("FindByID", ctx, userID, targetDeckID).Return(targetDeck, nil).Once()
		mockTM.ExpectTransaction()
		mockCardRepo.On("MoveCards", mock.Anything, userID, deckID, targetDeckID).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, userID, deckID).Return(nil).Once()

		err := service.Delete(ctx, userID, deckID, deck.ActionMoveToDeck, &targetDeckID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockBackupSvc.AssertExpectations(t)
	})

	t.Run("Prevent Default Deck Deletion", func(t *testing.T) {
		deckID := int64(1)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Default").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()

		err := service.Delete(ctx, userID, deckID, deck.ActionDeleteCards, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete the default deck")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deck Not Found", func(t *testing.T) {
		deckID := int64(404)
		mockRepo.On("FindByID", ctx, userID, deckID).Return(nil, nil).Once()

		err := service.Delete(ctx, userID, deckID, deck.ActionDeleteCards, nil)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, deckSvc.ErrDeckNotFound))
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckService_FindByUserID(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockBackupSvc := new(MockBackupService)
	mockTM := new(MockTransactionManager)
	service := deckSvc.NewDeckService(mockRepo, mockCardRepo, mockBackupSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		d1, _ := deck.NewBuilder().WithID(1).WithUserID(userID).WithName("Deck 1").Build()
		d2, _ := deck.NewBuilder().WithID(2).WithUserID(userID).WithName("Deck 2").Build()
		expectedDecks := []*deck.Deck{d1, d2}

		mockRepo.On("FindByUserID", ctx, userID).Return(expectedDecks, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedDecks, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty List", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return([]*deck.Deck{}, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckService_FindByID(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockBackupSvc := new(MockBackupService)
	mockTM := new(MockTransactionManager)
	service := deckSvc.NewDeckService(mockRepo, mockCardRepo, mockBackupSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		expectedDeck, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Target Deck").Build()
		mockRepo.On("FindByID", ctx, userID, deckID).Return(expectedDeck, nil).Once()

		result, err := service.FindByID(ctx, userID, deckID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedDeck, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deckID).Return(nil, nil).Once()

		result, err := service.FindByID(ctx, userID, deckID)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckService_Update(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockBackupSvc := new(MockBackupService)
	mockTM := new(MockTransactionManager)
	service := deckSvc.NewDeckService(mockRepo, mockCardRepo, mockBackupSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success Name Update", func(t *testing.T) {
		newName := "Updated Name"
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Old Name").Build()
		
		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Exists", ctx, userID, newName, (*int64)(nil)).Return(false, nil).Once()
		mockRepo.On("Update", ctx, userID, deckID, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, deckID, newName, nil, "")

		assert.NoError(t, err)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success Parent Update", func(t *testing.T) {
		name := "Deck Name"
		newParentID := int64(20)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName(name).Build()
		parentDeck, _ := deck.NewBuilder().WithID(newParentID).WithUserID(userID).WithName("Parent").Build()
		
		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Exists", ctx, userID, name, &newParentID).Return(false, nil).Once()
		mockRepo.On("FindByID", ctx, userID, newParentID).Return(parentDeck, nil).Once()
		mockRepo.On("Update", ctx, userID, deckID, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, deckID, name, &newParentID, "")

		assert.NoError(t, err)
		assert.Equal(t, &newParentID, result.GetParentID())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deck Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deckID).Return(nil, nil).Once()

		result, err := service.Update(ctx, userID, deckID, "Name", nil, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "deck not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Conflict Same Name At Same Level", func(t *testing.T) {
		newName := "Conflict Name"
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Old Name").Build()
		
		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Exists", ctx, userID, newName, (*int64)(nil)).Return(true, nil).Once()

		result, err := service.Update(ctx, userID, deckID, newName, nil, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deep Circular Dependency", func(t *testing.T) {
		name := "Name"
		childID := int64(20)
		grandchildID := int64(30)
		
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName(name).Build()
		child, _ := deck.NewBuilder().WithID(childID).WithUserID(userID).WithParentID(&deckID).WithName("Child").Build()
		grandchild, _ := deck.NewBuilder().WithID(grandchildID).WithUserID(userID).WithParentID(&childID).WithName("Grandchild").Build()
		
		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Exists", ctx, userID, name, &grandchildID).Return(false, nil).Once()
		
		// Traverse up from grandchild
		mockRepo.On("FindByID", ctx, userID, grandchildID).Return(grandchild, nil).Once()
		mockRepo.On("FindByID", ctx, userID, childID).Return(child, nil).Once()

		result, err := service.Update(ctx, userID, deckID, name, &grandchildID, "")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, deckSvc.ErrCircularDependency))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Circular Dependency Self", func(t *testing.T) {
		name := "Name"
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName(name).Build()
		
		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Exists", ctx, userID, name, &deckID).Return(false, nil).Once()

		result, err := service.Update(ctx, userID, deckID, name, &deckID, "")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, deckSvc.ErrCircularDependency))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Parent Not Found", func(t *testing.T) {
		name := "Name"
		newParentID := int64(20)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName(name).Build()
		
		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Exists", ctx, userID, name, &newParentID).Return(false, nil).Once()
		mockRepo.On("FindByID", ctx, userID, newParentID).Return(nil, nil).Once()

		result, err := service.Update(ctx, userID, deckID, name, &newParentID, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "parent deck not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckService_UpdateOptions(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockBackupSvc := new(MockBackupService)
	mockTM := new(MockTransactionManager)
	service := deckSvc.NewDeckService(mockRepo, mockCardRepo, mockBackupSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success", func(t *testing.T) {
		oldOptions := `{"old": "value"}`
		newOptions := `{"new": "value"}`
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithOptionsJSON(oldOptions).Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Update", ctx, userID, deckID, mock.AnythingOfType("*deck.Deck")).Return(nil).Once()

		result, err := service.UpdateOptions(ctx, userID, deckID, newOptions)

		assert.NoError(t, err)
		assert.Equal(t, newOptions, result.GetOptionsJSON())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deck Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deckID).Return(nil, nil).Once()

		result, err := service.UpdateOptions(ctx, userID, deckID, `{}`)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, deckSvc.ErrDeckNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

