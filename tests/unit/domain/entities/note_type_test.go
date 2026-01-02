package entities
import (
	"strings"
	"testing"
	"time"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
)

func TestNoteType_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		noteType *notetype.NoteType
		expected bool
	}{
		{
			name: "active note type",
			noteType: func() *notetype.NoteType {
				nt := &notetype.NoteType{}
				nt.SetDeletedAt(nil)
				return nt
			}(),
			expected: true,
		},
		{
			name: "deleted note type",
			noteType: func() *notetype.NoteType {
				nt := &notetype.NoteType{}
				nt.SetDeletedAt(timePtr(time.Now()))
				return nt
			}(),
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
			nt := &notetype.NoteType{}
			nt.SetFieldsJSON(tt.fieldsJSON)
			got := nt.GetFieldCount()
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
			nt := &notetype.NoteType{}
			nt.SetCardTypesJSON(tt.cardTypesJSON)
			got := nt.GetCardTypeCount()
			if got != tt.expected {
				t.Errorf("NoteType.GetCardTypeCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNoteType_GetFirstFieldName(t *testing.T) {
	tests := []struct {
		name        string
		fieldsJSON  string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "success - single field",
			fieldsJSON:  `[{"name": "Front", "ord": 0}]`,
			expected:    "Front",
			expectError: false,
		},
		{
			name:        "success - multiple fields",
			fieldsJSON:  `[{"name": "Question", "ord": 0}, {"name": "Answer", "ord": 1}]`,
			expected:    "Question",
			expectError: false,
		},
		{
			name:        "error - empty fields array",
			fieldsJSON:  `[]`,
			expectError: true,
			errorMsg:    "note type has no fields defined",
		},
		{
			name:        "error - empty string",
			fieldsJSON:  "",
			expectError: true,
			errorMsg:    "note type has no fields defined",
		},
		{
			name:        "error - invalid JSON",
			fieldsJSON:  "invalid json",
			expectError: true,
			errorMsg:    "invalid note type fields JSON",
		},
		{
			name:        "error - first field missing name",
			fieldsJSON:  `[{"ord": 0}, {"name": "Back"}]`,
			expectError: true,
			errorMsg:    "first field has no name property",
		},
		{
			name:        "error - first field name is empty string",
			fieldsJSON:  `[{"name": ""}, {"name": "Back"}]`,
			expectError: true,
			errorMsg:    "first field name is empty",
		},
		{
			name:        "error - first field name is not a string",
			fieldsJSON:  `[{"name": 123}, {"name": "Back"}]`,
			expectError: true,
			errorMsg:    "first field name is not a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nt := &notetype.NoteType{}
			nt.SetFieldsJSON(tt.fieldsJSON)
			got, err := nt.GetFirstFieldName()

			if tt.expectError {
				if err == nil {
					t.Errorf("NoteType.GetFirstFieldName() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("NoteType.GetFirstFieldName() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("NoteType.GetFirstFieldName() unexpected error = %v", err)
					return
				}
				if got != tt.expected {
					t.Errorf("NoteType.GetFirstFieldName() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}



