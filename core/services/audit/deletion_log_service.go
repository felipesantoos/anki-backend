package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// DeletionLogService implements IDeletionLogService
type DeletionLogService struct {
	repo      secondary.IDeletionLogRepository
	noteService primary.INoteService
	noteRepo  secondary.INoteRepository
}

// NewDeletionLogService creates a new DeletionLogService instance
func NewDeletionLogService(repo secondary.IDeletionLogRepository, noteService primary.INoteService, noteRepo secondary.INoteRepository) primary.IDeletionLogService {
	return &DeletionLogService{
		repo:       repo,
		noteService: noteService,
		noteRepo:   noteRepo,
	}
}

// Create records a new deletion event
func (s *DeletionLogService) Create(ctx context.Context, userID int64, objectType string, objectID int64) (*deletionlog.DeletionLog, error) {
	dl, err := deletionlog.NewBuilder().
		WithUserID(userID).
		WithObjectType(objectType).
		WithObjectID(objectID).
		WithDeletedAt(time.Now()).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, dl); err != nil {
		return nil, err
	}

	return dl, nil
}

// FindByUserID finds deletion logs for a user
func (s *DeletionLogService) FindByUserID(ctx context.Context, userID int64) ([]*deletionlog.DeletionLog, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// FindRecent finds recent deletion logs for a user within a specified time period
func (s *DeletionLogService) FindRecent(ctx context.Context, userID int64, limit int, days int) ([]*deletionlog.DeletionLog, error) {
	// Validate and set defaults
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	if days <= 0 {
		days = 7 // Default days
	}
	if days > 365 {
		days = 365 // Max days
	}

	return s.repo.FindRecent(ctx, userID, limit, days)
}

// Restore restores a deleted note from a deletion log entry
func (s *DeletionLogService) Restore(ctx context.Context, userID int64, deletionLogID int64, deckID int64) (*note.Note, error) {
	// 1. Fetch deletion log by ID (validates ownership)
	deletionLog, err := s.repo.FindByID(ctx, userID, deletionLogID)
	if err != nil {
		return nil, err
	}
	if deletionLog == nil {
		return nil, ownership.ErrResourceNotFound
	}

	// 2. Validate it can be recovered
	if !deletionLog.CanRecover() {
		return nil, fmt.Errorf("deletion log does not contain recoverable data")
	}

	// 3. Validate object_type is "note"
	if deletionLog.GetObjectType() != deletionlog.ObjectTypeNote {
		return nil, fmt.Errorf("can only restore notes, got object_type: %s", deletionLog.GetObjectType())
	}

	// 4. Parse object_data JSON
	var objectData map[string]interface{}
	if err := json.Unmarshal([]byte(deletionLog.GetObjectData()), &objectData); err != nil {
		return nil, fmt.Errorf("failed to parse object_data JSON: %w", err)
	}

	// Extract required fields
	guidStr, ok := objectData["guid"].(string)
	if !ok || guidStr == "" {
		return nil, fmt.Errorf("object_data missing or invalid guid")
	}

	noteTypeIDFloat, ok := objectData["note_type_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("object_data missing or invalid note_type_id")
	}
	noteTypeID := int64(noteTypeIDFloat)

	fields, ok := objectData["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("object_data missing or invalid fields")
	}
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields: %w", err)
	}

	var tags []string
	if tagsInterface, ok := objectData["tags"]; ok {
		if tagsArray, ok := tagsInterface.([]interface{}); ok {
			tags = make([]string, 0, len(tagsArray))
			for _, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}
	}

	// 5. Check if original GUID already exists (not deleted)
	originalGUID, err := valueobjects.NewGUID(guidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GUID in object_data: %w", err)
	}

	existingNote, err := s.noteRepo.FindByGUID(ctx, userID, originalGUID.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to check GUID conflict: %w", err)
	}

	// If GUID already exists and note is not deleted, it means the note was already restored
	if existingNote != nil {
		return nil, fmt.Errorf("note with GUID %s already exists and is not deleted - already restored", originalGUID.Value())
	}
	
	// Note: If the GUID exists but is soft-deleted, FindByGUID returns nil
	// The database UNIQUE constraint will prevent us from using the same GUID
	// This is handled in the Update step below

	// 6. Create note using noteService.Create (validates note type and deck)
	// Note: Create will generate a new GUID, so we'll update it with the original GUID after creation
	restoredNote, err := s.noteService.Create(ctx, userID, noteTypeID, deckID, string(fieldsJSON), tags)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// 7. Update note with original GUID
	restoredNote.SetGUID(originalGUID)
	if err := s.noteRepo.Update(ctx, userID, restoredNote.GetID(), restoredNote); err != nil {
		// Check if error is due to duplicate GUID (note already exists, possibly soft-deleted)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			// The GUID exists in the database (possibly soft-deleted)
			// This shouldn't happen in normal operation since we checked above
			// But it can happen if the note was soft-deleted
			// Delete the note we just created to clean up
			_ = s.noteRepo.Delete(ctx, userID, restoredNote.GetID())
			return nil, fmt.Errorf("note with GUID %s already exists - already restored", originalGUID.Value())
		}
		// If update fails for other reasons, we still have the note created, so return error
		// The note will have a new GUID which is not ideal but acceptable
		return nil, fmt.Errorf("note restored but failed to set original GUID: %w", err)
	}

	return restoredNote, nil
}

