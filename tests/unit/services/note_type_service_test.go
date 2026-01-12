package services

import (
	"context"
	"testing"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	noteTypeSvc "github.com/felipesantos/anki-backend/core/services/notetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteTypeService_Create(t *testing.T) {
	mockRepo := new(MockNoteTypeRepository)
	service := noteTypeSvc.NewNoteTypeService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "Basic"
		mockRepo.On("ExistsByName", ctx, userID, name).Return(false, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*notetype.NoteType")).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", "[]", "{}")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Already Exists", func(t *testing.T) {
		name := "Duplicate"
		mockRepo.On("ExistsByName", ctx, userID, name).Return(true, nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", "[]", "{}")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rejects empty front template", func(t *testing.T) {
		name := "Empty Template"
		cardTypesJSON := `[{"name": "Card 1"}]`
		templatesJSON := `[{"qfmt": "", "afmt": "{{Back}}"}]`

		mockRepo.On("ExistsByName", ctx, userID, name).Return(false, nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", cardTypesJSON, templatesJSON)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "front template (qfmt) cannot be empty")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rejects missing qfmt field", func(t *testing.T) {
		name := "Missing qfmt"
		cardTypesJSON := `[{"name": "Card 1"}]`
		templatesJSON := `[{"afmt": "{{Back}}"}]`

		mockRepo.On("ExistsByName", ctx, userID, name).Return(false, nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", cardTypesJSON, templatesJSON)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "front template (qfmt) missing")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rejects missing template for card type", func(t *testing.T) {
		name := "Missing Template"
		cardTypesJSON := `[{"name": "Card 1"}, {"name": "Card 2"}]`
		templatesJSON := `[{"qfmt": "{{Front}}", "afmt": "{{Back}}"}]` // Only 1 template for 2 card types

		mockRepo.On("ExistsByName", ctx, userID, name).Return(false, nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", cardTypesJSON, templatesJSON)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template missing for card type 1")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNoteTypeService_Update(t *testing.T) {
	mockRepo := new(MockNoteTypeRepository)
	service := noteTypeSvc.NewNoteTypeService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	id := int64(100)

	t.Run("Success", func(t *testing.T) {
		existing, _ := notetype.NewBuilder().WithID(id).WithUserID(userID).WithName("Old Name").Build()
		newName := "New Name"
		cardTypesJSON := `[{"name": "Card 1"}]`
		templatesJSON := `[{"qfmt": "{{Front}}", "afmt": "{{Back}}"}]`

		mockRepo.On("FindByID", ctx, userID, id).Return(existing, nil).Once()
		mockRepo.On("ExistsByName", ctx, userID, newName).Return(false, nil).Once()
		mockRepo.On("Update", ctx, userID, id, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, id, newName, "[]", cardTypesJSON, templatesJSON)

		assert.NoError(t, err)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rejects empty front template", func(t *testing.T) {
		existing, _ := notetype.NewBuilder().WithID(id).WithUserID(userID).WithName("Old Name").Build()
		newName := "Old Name"
		cardTypesJSON := `[{"name": "Card 1"}]`
		templatesJSON := `[{"qfmt": "", "afmt": "{{Back}}"}]`

		mockRepo.On("FindByID", ctx, userID, id).Return(existing, nil).Once()

		result, err := service.Update(ctx, userID, id, newName, "[]", cardTypesJSON, templatesJSON)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "front template (qfmt) cannot be empty")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rejects missing template for card type", func(t *testing.T) {
		existing, _ := notetype.NewBuilder().WithID(id).WithUserID(userID).WithName("Old Name").Build()
		newName := "Old Name"
		cardTypesJSON := `[{"name": "Card 1"}, {"name": "Card 2"}]`
		templatesJSON := `[{"qfmt": "{{Front}}", "afmt": "{{Back}}"}]`

		mockRepo.On("FindByID", ctx, userID, id).Return(existing, nil).Once()

		result, err := service.Update(ctx, userID, id, newName, "[]", cardTypesJSON, templatesJSON)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template missing for card type 1")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNoteTypeService_FindByUserID(t *testing.T) {
	mockRepo := new(MockNoteTypeRepository)
	service := noteTypeSvc.NewNoteTypeService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success with Search", func(t *testing.T) {
		search := "Basic"
		nt1, _ := notetype.NewBuilder().WithID(1).WithUserID(userID).WithName("Basic").Build()
		expectedNoteTypes := []*notetype.NoteType{nt1}

		mockRepo.On("FindByUserID", ctx, userID, search).Return(expectedNoteTypes, nil).Once()

		result, err := service.FindByUserID(ctx, userID, search)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expectedNoteTypes, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Case-Insensitive Search", func(t *testing.T) {
		search := "basic" // lowercase search for "Basic"
		nt1, _ := notetype.NewBuilder().WithID(1).WithUserID(userID).WithName("Basic").Build()
		expectedNoteTypes := []*notetype.NoteType{nt1}

		mockRepo.On("FindByUserID", ctx, userID, search).Return(expectedNoteTypes, nil).Once()

		result, err := service.FindByUserID(ctx, userID, search)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expectedNoteTypes, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Partial Match Search", func(t *testing.T) {
		search := "Basic"
		nt1, _ := notetype.NewBuilder().WithID(1).WithUserID(userID).WithName("Basic").Build()
		nt2, _ := notetype.NewBuilder().WithID(2).WithUserID(userID).WithName("Basic with Reverso").Build()
		expectedNoteTypes := []*notetype.NoteType{nt1, nt2}

		mockRepo.On("FindByUserID", ctx, userID, search).Return(expectedNoteTypes, nil).Once()

		result, err := service.FindByUserID(ctx, userID, search)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedNoteTypes, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("No Matches", func(t *testing.T) {
		search := "NonExistent"
		mockRepo.On("FindByUserID", ctx, userID, search).Return([]*notetype.NoteType{}, nil).Once()

		result, err := service.FindByUserID(ctx, userID, search)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty Search String", func(t *testing.T) {
		nt1, _ := notetype.NewBuilder().WithID(1).WithUserID(userID).WithName("Basic").Build()
		nt2, _ := notetype.NewBuilder().WithID(2).WithUserID(userID).WithName("Cloze").Build()
		expectedNoteTypes := []*notetype.NoteType{nt1, nt2}

		mockRepo.On("FindByUserID", ctx, userID, "").Return(expectedNoteTypes, nil).Once()

		result, err := service.FindByUserID(ctx, userID, "")

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedNoteTypes, result)
		mockRepo.AssertExpectations(t)
	})
}

