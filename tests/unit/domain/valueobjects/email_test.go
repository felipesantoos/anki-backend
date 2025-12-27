package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrEmailEmpty  = valueobjects.ErrEmailEmpty
	ErrEmailInvalid = valueobjects.ErrEmailInvalid
	NewEmail       = valueobjects.NewEmail
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid email",
			input:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with uppercase",
			input:   "USER@EXAMPLE.COM",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			input:   "user123@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with special characters",
			input:   "user.name+tag@example.co.uk",
			wantErr: false,
		},
		{
			name:    "empty email",
			input:   "",
			wantErr: true,
			errType: ErrEmailEmpty,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
			errType: ErrEmailEmpty,
		},
		{
			name:    "invalid format - no @",
			input:   "userexample.com",
			wantErr: true,
			errType: ErrEmailInvalid,
		},
		{
			name:    "invalid format - no domain",
			input:   "user@",
			wantErr: true,
			errType: ErrEmailInvalid,
		},
		{
			name:    "invalid format - no TLD",
			input:   "user@example",
			wantErr: true,
			errType: ErrEmailInvalid,
		},
		{
			name:    "invalid format - spaces",
			input:   "user @example.com",
			wantErr: true,
			errType: ErrEmailInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEmail() expected error, got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewEmail() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewEmail() unexpected error = %v", err)
				return
			}

			// Verify email is lowercase
			value := email.Value()
			if value != tt.input && value != toLowerAndTrim(tt.input) {
				// Email should be normalized to lowercase
				expected := toLowerAndTrim(tt.input)
				if value != expected {
					t.Errorf("NewEmail() value = %v, want normalized version", value)
				}
			}
		})
	}
}

func TestEmail_Value(t *testing.T) {
	email, err := NewEmail("User@Example.COM")
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}

	value := email.Value()
	expected := "user@example.com"
	if value != expected {
		t.Errorf("Email.Value() = %v, want %v", value, expected)
	}
}

func TestEmail_Equals(t *testing.T) {
	email1, err := NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}

	email2, err := NewEmail("USER@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}

	email3, err := NewEmail("other@example.com")
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}

	if !email1.Equals(email2) {
		t.Errorf("Email.Equals() should return true for same email (case-insensitive)")
	}

	if email1.Equals(email3) {
		t.Errorf("Email.Equals() should return false for different emails")
	}
}

func TestEmail_String(t *testing.T) {
	email, err := NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}

	str := email.String()
	expected := "user@example.com"
	if str != expected {
		t.Errorf("Email.String() = %v, want %v", str, expected)
	}
}

// Helper function to normalize email (lowercase and trim)
func toLowerAndTrim(s string) string {
	// Simple implementation for testing
	// In real code, NewEmail does this internally
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32) // Convert to lowercase
		} else {
			result += string(r)
		}
	}
	// Trim spaces (simplified)
	for len(result) > 0 && result[0] == ' ' {
		result = result[1:]
	}
	for len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return result
}
