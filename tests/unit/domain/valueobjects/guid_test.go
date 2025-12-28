package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestNewGUID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid UUID lowercase",
			input:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid UUID uppercase",
			input:   "550E8400-E29B-41D4-A716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid UUID mixed case",
			input:   "550E8400-e29b-41D4-A716-446655440000",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errType: valueobjects.ErrGUIDEmpty,
		},
		{
			name:    "invalid format - missing hyphens",
			input:   "550e8400e29b41d4a716446655440000",
			wantErr: true,
			errType: valueobjects.ErrGUIDInvalid,
		},
		{
			name:    "invalid format - wrong length",
			input:   "550e8400-e29b-41d4-a716-44665544",
			wantErr: true,
			errType: valueobjects.ErrGUIDInvalid,
		},
		{
			name:    "invalid format - invalid characters",
			input:   "550e8400-e29b-41d4-a716-44665544000g",
			wantErr: true,
			errType: valueobjects.ErrGUIDInvalid,
		},
		{
			name:    "invalid format - spaces",
			input:   "550e8400-e29b-41d4-a716-44665544000 ",
			wantErr: true,
			errType: valueobjects.ErrGUIDInvalid,
		},
	}

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guid, err := valueobjects.NewGUID(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewGUID() expected error, got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewGUID() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewGUID() unexpected error = %v", err)
				return
			}

			// Verify GUID is normalized to lowercase
			if guid.Value() != "550e8400-e29b-41d4-a716-446655440000" {
				t.Errorf("NewGUID() value = %v, want normalized lowercase", guid.Value())
			}
		})
	}
}

func TestGUID_Value(t *testing.T) {
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewGUID() error = %v", err)
	}

	if guid.Value() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("GUID.Value() = %v, want '550e8400-e29b-41d4-a716-446655440000'", guid.Value())
	}
}

func TestGUID_String(t *testing.T) {
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewGUID() error = %v", err)
	}

	if guid.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("GUID.String() = %v, want '550e8400-e29b-41d4-a716-446655440000'", guid.String())
	}
}

func TestGUID_Equals(t *testing.T) {
	guid1, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewGUID() error = %v", err)
	}

	guid2, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewGUID() error = %v", err)
	}

	guid3, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
	if err != nil {
		t.Fatalf("NewGUID() error = %v", err)
	}

	if !guid1.Equals(guid2) {
		t.Errorf("GUID.Equals() = false, want true for same GUIDs")
	}

	if guid1.Equals(guid3) {
		t.Errorf("GUID.Equals() = true, want false for different GUIDs")
	}
}

func TestGUID_IsEmpty(t *testing.T) {
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewGUID() error = %v", err)
	}

	if guid.IsEmpty() {
		t.Errorf("GUID.IsEmpty() = true, want false for valid GUID")
	}

	emptyGUID := valueobjects.GUID{}
	if !emptyGUID.IsEmpty() {
		t.Errorf("GUID.IsEmpty() = false, want true for empty GUID")
	}
}

func TestNewGUIDFromString(t *testing.T) {
	guid, err := valueobjects.NewGUIDFromString("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewGUIDFromString() error = %v", err)
	}

	if guid.Value() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("NewGUIDFromString() value = %v, want '550e8400-e29b-41d4-a716-446655440000'", guid.Value())
	}
}

