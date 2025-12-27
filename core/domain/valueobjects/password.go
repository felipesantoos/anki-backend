package valueobjects

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordTooShort is returned when password has less than 8 characters
	ErrPasswordTooShort = errors.New("password must have at least 8 characters")
	// ErrPasswordNoLetter is returned when password doesn't contain at least one letter
	ErrPasswordNoLetter = errors.New("password must contain at least one letter")
	// ErrPasswordNoNumber is returned when password doesn't contain at least one number
	ErrPasswordNoNumber = errors.New("password must contain at least one number")
)

const (
	// MinPasswordLength is the minimum required length for passwords
	MinPasswordLength = 8
	// bcryptCost is the cost factor for bcrypt hashing (higher = more secure but slower)
	bcryptCost = bcrypt.DefaultCost
)

// Password represents a validated password value object
// It provides validation and hashing functionality
type Password struct {
	hashed string
}

// NewPassword creates a new Password value object from a plain text password
// It validates the password strength according to business rules:
// - Minimum 8 characters
// - At least one letter
// - At least one number
// Returns an error if validation fails
func NewPassword(plainText string) (Password, error) {
	if err := validatePasswordStrength(plainText); err != nil {
		return Password{}, err
	}

	hashed, err := hashPassword(plainText)
	if err != nil {
		return Password{}, err
	}

	return Password{hashed: hashed}, nil
}

// NewPasswordFromHash creates a Password from an already hashed password
// This is useful when loading passwords from the database
func NewPasswordFromHash(hashed string) Password {
	return Password{hashed: hashed}
}

// Hash returns the bcrypt hash of the password
func (p Password) Hash() string {
	return p.hashed
}

// Verify compares a plain text password with the hashed password
// Returns true if they match, false otherwise
func (p Password) Verify(plainText string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hashed), []byte(plainText))
	return err == nil
}

// validatePasswordStrength validates password according to business rules
func validatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
	}

	if !hasLetter {
		return ErrPasswordNoLetter
	}

	if !hasNumber {
		return ErrPasswordNoNumber
	}

	return nil
}

// hashPassword hashes a plain text password using bcrypt
func hashPassword(plainText string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
