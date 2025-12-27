package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrPasswordTooShort = valueobjects.ErrPasswordTooShort
	ErrPasswordNoLetter = valueobjects.ErrPasswordNoLetter
	ErrPasswordNoNumber = valueobjects.ErrPasswordNoNumber
	NewPassword         = valueobjects.NewPassword
	NewPasswordFromHash = valueobjects.NewPasswordFromHash
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid password",
			input:   "password123",
			wantErr: false,
		},
		{
			name:    "valid password with uppercase",
			input:   "Password123",
			wantErr: false,
		},
		{
			name:    "valid password with special characters",
			input:   "Pass@word123",
			wantErr: false,
		},
		{
			name:    "password too short",
			input:   "pass123",
			wantErr: true,
			errType: ErrPasswordTooShort,
		},
		{
			name:    "password with no letters",
			input:   "12345678",
			wantErr: true,
			errType: ErrPasswordNoLetter,
		},
		{
			name:    "password with no numbers",
			input:   "password",
			wantErr: true,
			errType: ErrPasswordNoNumber,
		},
		{
			name:    "minimum length password with letter and number",
			input:   "pass1234",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := NewPassword(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPassword() expected error, got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewPassword() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewPassword() unexpected error = %v", err)
				return
			}

			// Verify password is hashed (not plain text)
			hash := password.Hash()
			if hash == tt.input {
				t.Errorf("NewPassword() hash should not equal plain text password")
			}
			if len(hash) == 0 {
				t.Errorf("NewPassword() hash should not be empty")
			}
		})
	}
}

func TestPassword_Verify(t *testing.T) {
	plainText := "password123"
	password, err := NewPassword(plainText)
	if err != nil {
		t.Fatalf("NewPassword() error = %v", err)
	}

	// Verify correct password
	if !password.Verify(plainText) {
		t.Errorf("Password.Verify() should return true for correct password")
	}

	// Verify incorrect password
	if password.Verify("wrongpassword") {
		t.Errorf("Password.Verify() should return false for incorrect password")
	}

	// Verify empty password
	if password.Verify("") {
		t.Errorf("Password.Verify() should return false for empty password")
	}
}

func TestNewPasswordFromHash(t *testing.T) {
	// Create a password first to get a valid hash
	originalPassword, err := NewPassword("password123")
	if err != nil {
		t.Fatalf("NewPassword() error = %v", err)
	}

	hash := originalPassword.Hash()

	// Create password from hash
	passwordFromHash := NewPasswordFromHash(hash)

	// Verify it works with the original password
	if !passwordFromHash.Verify("password123") {
		t.Errorf("Password created from hash should verify correctly")
	}

	// Verify it doesn't work with wrong password
	if passwordFromHash.Verify("wrongpassword") {
		t.Errorf("Password created from hash should not verify incorrect password")
	}
}

func TestPassword_Hash(t *testing.T) {
	password, err := NewPassword("password123")
	if err != nil {
		t.Fatalf("NewPassword() error = %v", err)
	}

	hash := password.Hash()
	if len(hash) == 0 {
		t.Errorf("Password.Hash() should not be empty")
	}

	// Hash should start with bcrypt identifier
	if len(hash) < 60 {
		t.Errorf("Password.Hash() should be at least 60 characters (bcrypt hash)")
	}
}
