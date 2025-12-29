package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestNote_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		note     *note.Note
		expected bool
	}{
		{
			name: "active note",
			note: func() *note.Note {
				guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
				n, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithDeletedAt(nil).Build()
				return n
			}(),
			expected: true,
		},
		{
			name: "deleted note",
			note: func() *note.Note {
				guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
				now := time.Now()
				n, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithDeletedAt(&now).Build()
				return n
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.note.IsActive()
			if got != tt.expected {
				t.Errorf("Note.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNote_HasTag(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	noteEntity, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithTags([]string{"vocabulary", "spanish", "verb"}).Build()

	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{
			name:     "has tag",
			tag:      "vocabulary",
			expected: true,
		},
		{
			name:     "has tag case insensitive",
			tag:      "VOCABULARY",
			expected: true,
		},
		{
			name:     "does not have tag",
			tag:      "noun",
			expected: false,
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := noteEntity.HasTag(tt.tag)
			if got != tt.expected {
				t.Errorf("Note.HasTag(%q) = %v, want %v", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestNote_AddTag(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	noteEntity, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithTags([]string{"vocabulary"}).WithUpdatedAt(time.Now()).Build()

	// Add new tag
	originalUpdatedAt := noteEntity.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)
	noteEntity.AddTag("spanish")
	if !noteEntity.HasTag("spanish") {
		t.Errorf("Note.AddTag() failed to add tag")
	}

	// Verify UpdatedAt was changed when adding new tag
	if noteEntity.GetUpdatedAt().Equal(originalUpdatedAt) {
		t.Errorf("Note.AddTag() should update UpdatedAt when adding new tag")
	}

	// Try to add duplicate tag (should not add duplicate)
	noteEntity.AddTag("spanish")
	if len(noteEntity.GetTags()) != 2 {
		t.Errorf("Note.AddTag() added duplicate tag, want 2 tags, got %d", len(noteEntity.GetTags()))
	}

	// Try to add empty tag
	originalTagCount := len(noteEntity.GetTags())
	noteEntity.AddTag("")
	if len(noteEntity.GetTags()) != originalTagCount {
		t.Errorf("Note.AddTag() should not add empty tag")
	}
}

func TestNote_RemoveTag(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	noteEntity, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithTags([]string{"vocabulary", "spanish", "verb"}).WithUpdatedAt(time.Now()).Build()

	// Remove existing tag
	noteEntity.RemoveTag("spanish")
	if noteEntity.HasTag("spanish") {
		t.Errorf("Note.RemoveTag() failed to remove tag")
	}

	if len(noteEntity.GetTags()) != 2 {
		t.Errorf("Note.RemoveTag() wrong tag count, want 2, got %d", len(noteEntity.GetTags()))
	}

	// Try to remove non-existent tag
	originalTagCount := len(noteEntity.GetTags())
	noteEntity.RemoveTag("nonexistent")
	if len(noteEntity.GetTags()) != originalTagCount {
		t.Errorf("Note.RemoveTag() should not change tag count for non-existent tag")
	}

	// Remove tag case insensitive
	noteEntity.RemoveTag("VOCABULARY")
	if noteEntity.HasTag("vocabulary") {
		t.Errorf("Note.RemoveTag() should be case insensitive")
	}
}

func TestNote_IsMarked(t *testing.T) {
	guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	markedNote, _ := note.NewBuilder().WithUserID(1).WithGUID(guid1).WithNoteTypeID(1).WithMarked(true).Build()
	if !markedNote.IsMarked() {
		t.Errorf("Note.IsMarked() = false, want true for marked note")
	}

	guid2, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
	unmarkedNote, _ := note.NewBuilder().WithUserID(1).WithGUID(guid2).WithNoteTypeID(1).WithMarked(false).Build()
	if unmarkedNote.IsMarked() {
		t.Errorf("Note.IsMarked() = true, want false for unmarked note")
	}
}

func TestNote_Mark(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	noteEntity, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithMarked(false).WithUpdatedAt(time.Now()).Build()

	noteEntity.Mark()
	if !noteEntity.GetMarked() {
		t.Errorf("Note.Mark() failed to mark note")
	}

	// Mark again (should be idempotent)
	noteEntity.Mark()
	if !noteEntity.GetMarked() {
		t.Errorf("Note.Mark() failed to keep note marked")
	}
}

func TestNote_Unmark(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	noteEntity, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithMarked(true).WithUpdatedAt(time.Now()).Build()

	noteEntity.Unmark()
	if noteEntity.GetMarked() {
		t.Errorf("Note.Unmark() failed to unmark note")
	}

	// Unmark again (should be idempotent)
	noteEntity.Unmark()
	if noteEntity.GetMarked() {
		t.Errorf("Note.Unmark() failed to keep note unmarked")
	}
}

func TestNote_GetFirstField(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	noteEntity, _ := note.NewBuilder().WithUserID(1).WithGUID(guid).WithNoteTypeID(1).WithFieldsJSON(`{"Front": "Hello", "Back": "Hola"}`).Build()

	// GetFirstField returns empty string as parsing should be done in service layer
	// This is expected behavior per the implementation
	result := noteEntity.GetFirstField()
	if result != "" {
		t.Errorf("Note.GetFirstField() = %v, want empty string (parsing in service layer)", result)
	}
}
