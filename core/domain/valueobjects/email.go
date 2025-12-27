package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrEmailEmpty is returned when email is empty
	ErrEmailEmpty = errors.New("email cannot be empty")
	// ErrEmailInvalid is returned when email format is invalid
	ErrEmailInvalid = errors.New("email format is invalid")
)

// emailRegex validates email format
// This regex pattern is based on RFC 5322 but simplified for practical use
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email represents a validated email address value object
// It is immutable and ensures the email is always in a valid format
type Email struct {
	value string
}

// NewEmail creates a new Email value object
// It validates the email format and normalizes it (lowercase, trimmed)
// Returns an error if the email is invalid
func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Email{}, ErrEmailEmpty
	}

	if !emailRegex.MatchString(trimmed) {
		return Email{}, ErrEmailInvalid
	}

	return Email{value: trimmed}, nil
}

// Value returns the email value as a string
func (e Email) Value() string {
	return e.value
}

// Equals compares two Email value objects for equality
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// String returns the string representation of the email
func (e Email) String() string {
	return e.value
}
