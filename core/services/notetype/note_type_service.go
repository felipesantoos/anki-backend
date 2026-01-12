package notetype

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// NoteTypeService implements INoteTypeService
type NoteTypeService struct {
	noteTypeRepo secondary.INoteTypeRepository
}

// NewNoteTypeService creates a new NoteTypeService instance
func NewNoteTypeService(noteTypeRepo secondary.INoteTypeRepository) primary.INoteTypeService {
	return &NoteTypeService{
		noteTypeRepo: noteTypeRepo,
	}
}

// Create creates a new note type
func (s *NoteTypeService) Create(ctx context.Context, userID int64, name string, fieldsJSON string, cardTypesJSON string, templatesJSON string) (*notetype.NoteType, error) {
	// 1. Check if note type with same name exists for user
	exists, err := s.noteTypeRepo.ExistsByName(ctx, userID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check note type existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("note type with name %s already exists", name)
	}

	// 1.5. Validate front templates are not empty
	cardTypesCount, err := s.getCardTypesCount(cardTypesJSON)
	if err != nil {
		return nil, err
	}
	if err := s.validateFrontTemplates(templatesJSON, cardTypesCount); err != nil {
		return nil, err
	}

	// 2. Create note type entity using builder
	now := time.Now()
	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName(name).
		WithFieldsJSON(fieldsJSON).
		WithCardTypesJSON(cardTypesJSON).
		WithTemplatesJSON(templatesJSON).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build note type entity: %w", err)
	}

	// 3. Save to repository
	if err := s.noteTypeRepo.Save(ctx, userID, noteTypeEntity); err != nil {
		return nil, fmt.Errorf("failed to save note type: %w", err)
	}

	return noteTypeEntity, nil
}

// FindByID finds a note type by ID
func (s *NoteTypeService) FindByID(ctx context.Context, userID int64, id int64) (*notetype.NoteType, error) {
	return s.noteTypeRepo.FindByID(ctx, userID, id)
}

// FindByUserID finds all note types for a user, with optional search filter
func (s *NoteTypeService) FindByUserID(ctx context.Context, userID int64, search string) ([]*notetype.NoteType, error) {
	return s.noteTypeRepo.FindByUserID(ctx, userID, search)
}

// Update updates an existing note type
func (s *NoteTypeService) Update(ctx context.Context, userID int64, id int64, name string, fieldsJSON string, cardTypesJSON string, templatesJSON string) (*notetype.NoteType, error) {
	// 1. Find existing note type
	existing, err := s.noteTypeRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("note type not found")
	}

	// 1.5. Validate front templates are not empty
	cardTypesCount, err := s.getCardTypesCount(cardTypesJSON)
	if err != nil {
		return nil, err
	}
	if err := s.validateFrontTemplates(templatesJSON, cardTypesCount); err != nil {
		return nil, err
	}

	// 2. If name changed, check for conflicts
	if existing.GetName() != name {
		exists, err := s.noteTypeRepo.ExistsByName(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("note type with name %s already exists", name)
		}
	}

	// 3. Update entity
	existing.SetName(name)
	existing.SetFieldsJSON(fieldsJSON)
	existing.SetCardTypesJSON(cardTypesJSON)
	existing.SetTemplatesJSON(templatesJSON)
	existing.SetUpdatedAt(time.Now())

	// 4. Save
	if err := s.noteTypeRepo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a note type (soft delete)
func (s *NoteTypeService) Delete(ctx context.Context, userID int64, id int64) error {
	// 1. Check if note type exists
	existing, err := s.noteTypeRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("note type not found")
	}

	// 2. Perform soft delete
	return s.noteTypeRepo.Delete(ctx, userID, id)
}

// validateFrontTemplates validates that all templates have non-empty qfmt fields
func (s *NoteTypeService) validateFrontTemplates(templatesJSON string, cardTypesCount int) error {
	// 1. Parse templatesJSON
	// It can be an array of objects or a single object (fallback)
	var templates []map[string]interface{}
	if err := json.Unmarshal([]byte(templatesJSON), &templates); err != nil {
		// Try parsing as single object if array fails
		var singleTemplate map[string]interface{}
		if err2 := json.Unmarshal([]byte(templatesJSON), &singleTemplate); err2 == nil {
			templates = []map[string]interface{}{singleTemplate}
		} else {
			return fmt.Errorf("invalid templates JSON: %w", err)
		}
	}

	// 2. Validate each template has non-empty qfmt
	for i := 0; i < cardTypesCount; i++ {
		if i >= len(templates) {
			return fmt.Errorf("template missing for card type %d", i)
		}

		template := templates[i]
		qfmt, exists := template["qfmt"]
		if !exists {
			// Try "Front" as fallback if qfmt doesn't exist
			qfmt, exists = template["Front"]
			if !exists {
				return fmt.Errorf("front template (qfmt) missing for card type %d", i)
			}
		}

		qfmtStr, ok := qfmt.(string)
		if !ok {
			return fmt.Errorf("front template (qfmt) must be a string for card type %d", i)
		}

		if strings.TrimSpace(qfmtStr) == "" {
			return fmt.Errorf("front template (qfmt) cannot be empty for card type %d", i)
		}
	}

	return nil
}

// getCardTypesCount parses cardTypesJSON and returns the number of card types
func (s *NoteTypeService) getCardTypesCount(cardTypesJSON string) (int, error) {
	if cardTypesJSON == "" {
		return 0, nil
	}

	var cardTypes []interface{}
	if err := json.Unmarshal([]byte(cardTypesJSON), &cardTypes); err != nil {
		return 0, fmt.Errorf("invalid card types JSON: %w", err)
	}

	return len(cardTypes), nil
}
