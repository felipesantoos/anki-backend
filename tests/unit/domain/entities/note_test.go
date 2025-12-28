package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestNote_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		note     *entities.Note
		expected bool
	}{
		{
			name: "active note",
			note: &entities.Note{
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "deleted note",
			note: &entities.Note{
				DeletedAt: timePtr(time.Now()),
			},
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
	note := &entities.Note{
		Tags: []string{"vocabulary", "spanish", "verb"},
	}

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
			got := note.HasTag(tt.tag)
			if got != tt.expected {
				t.Errorf("Note.HasTag(%q) = %v, want %v", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestNote_AddTag(t *testing.T) {
	note := &entities.Note{
		Tags:      []string{"vocabulary"},
		UpdatedAt: time.Now(),
	}

	// Add new tag
	originalUpdatedAt := note.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	note.AddTag("spanish")
	if !note.HasTag("spanish") {
		t.Errorf("Note.AddTag() failed to add tag")
	}

	// Verify UpdatedAt was changed when adding new tag
	if note.UpdatedAt.Equal(originalUpdatedAt) {
		t.Errorf("Note.AddTag() should update UpdatedAt when adding new tag")
	}

	// Try to add duplicate tag (should not add duplicate)
	note.AddTag("spanish")
	if len(note.Tags) != 2 {
		t.Errorf("Note.AddTag() added duplicate tag, want 2 tags, got %d", len(note.Tags))
	}

	// Try to add empty tag
	originalTagCount := len(note.Tags)
	note.AddTag("")
	if len(note.Tags) != originalTagCount {
		t.Errorf("Note.AddTag() should not add empty tag")
	}
}

func TestNote_RemoveTag(t *testing.T) {
	note := &entities.Note{
		Tags:      []string{"vocabulary", "spanish", "verb"},
		UpdatedAt: time.Now(),
	}

	// Remove existing tag
	note.RemoveTag("spanish")
	if note.HasTag("spanish") {
		t.Errorf("Note.RemoveTag() failed to remove tag")
	}

	if len(note.Tags) != 2 {
		t.Errorf("Note.RemoveTag() wrong tag count, want 2, got %d", len(note.Tags))
	}

	// Try to remove non-existent tag
	originalTagCount := len(note.Tags)
	note.RemoveTag("nonexistent")
	if len(note.Tags) != originalTagCount {
		t.Errorf("Note.RemoveTag() should not change tag count for non-existent tag")
	}

	// Remove tag case insensitive
	note.RemoveTag("VOCABULARY")
	if note.HasTag("vocabulary") {
		t.Errorf("Note.RemoveTag() should be case insensitive")
	}
}

func TestNote_IsMarked(t *testing.T) {
	markedNote := &entities.Note{Marked: true}
	if !markedNote.IsMarked() {
		t.Errorf("Note.IsMarked() = false, want true for marked note")
	}

	unmarkedNote := &entities.Note{Marked: false}
	if unmarkedNote.IsMarked() {
		t.Errorf("Note.IsMarked() = true, want false for unmarked note")
	}
}

func TestNote_Mark(t *testing.T) {
	note := &entities.Note{
		Marked:    false,
		UpdatedAt: time.Now(),
	}

	note.Mark()
	if !note.Marked {
		t.Errorf("Note.Mark() failed to mark note")
	}

	// Mark again (should be idempotent)
	note.Mark()
	if !note.Marked {
		t.Errorf("Note.Mark() failed to keep note marked")
	}
}

func TestNote_Unmark(t *testing.T) {
	note := &entities.Note{
		Marked:    true,
		UpdatedAt: time.Now(),
	}

	note.Unmark()
	if note.Marked {
		t.Errorf("Note.Unmark() failed to unmark note")
	}

	// Unmark again (should be idempotent)
	note.Unmark()
	if note.Marked {
		t.Errorf("Note.Unmark() failed to keep note unmarked")
	}
}

func TestNote_GetFirstField(t *testing.T) {
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	note := &entities.Note{
		GUID:       guid,
		FieldsJSON: `{"Front": "Hello", "Back": "Hola"}`,
	}

	// GetFirstField returns empty string as parsing should be done in service layer
	// This is expected behavior per the implementation
	result := note.GetFirstField()
	if result != "" {
		t.Errorf("Note.GetFirstField() = %v, want empty string (parsing in service layer)", result)
	}
}


