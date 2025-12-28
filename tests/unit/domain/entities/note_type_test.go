package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

import (
	"testing"
	"time"
)

func TestNoteType_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		noteType *entities.NoteType
		expected bool
	}{
		{
			name: "active note type",
			noteType: &entities.NoteType{
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "deleted note type",
			noteType: &entities.NoteType{
				DeletedAt: timePtr(time.Now()),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.noteType.IsActive()
			if got != tt.expected {
				t.Errorf("NoteType.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNoteType_GetFieldCount(t *testing.T) {
	tests := []struct {
		name       string
		fieldsJSON string
		expected   int
	}{
		{
			name:       "empty JSON array",
			fieldsJSON: "[]",
			expected:   0,
		},
		{
			name:       "single field",
			fieldsJSON: `[{"name": "Front", "ord": 0}]`,
			expected:   1,
		},
		{
			name:       "multiple fields",
			fieldsJSON: `[{"name": "Front", "ord": 0}, {"name": "Back", "ord": 1}]`,
			expected:   2,
		},
		{
			name:       "empty string",
			fieldsJSON: "",
			expected:   0,
		},
		{
			name:       "invalid JSON",
			fieldsJSON: "invalid json",
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noteType := &entities.NoteType{
				FieldsJSON: tt.fieldsJSON,
			}
			got := noteType.GetFieldCount()
			if got != tt.expected {
				t.Errorf("NoteType.GetFieldCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNoteType_GetCardTypeCount(t *testing.T) {
	tests := []struct {
		name          string
		cardTypesJSON string
		expected      int
	}{
		{
			name:          "empty JSON array",
			cardTypesJSON: "[]",
			expected:      0,
		},
		{
			name:          "single card type",
			cardTypesJSON: `[{"name": "Forward", "ord": 0}]`,
			expected:      1,
		},
		{
			name:          "multiple card types",
			cardTypesJSON: `[{"name": "Forward", "ord": 0}, {"name": "Reverse", "ord": 1}]`,
			expected:      2,
		},
		{
			name:          "empty string",
			cardTypesJSON: "",
			expected:      0,
		},
		{
			name:          "invalid JSON",
			cardTypesJSON: "invalid json",
			expected:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noteType := &entities.NoteType{
				CardTypesJSON: tt.cardTypesJSON,
			}
			got := noteType.GetCardTypeCount()
			if got != tt.expected {
				t.Errorf("NoteType.GetCardTypeCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}


