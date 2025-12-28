package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrGUIDEmpty is returned when GUID is empty
	ErrGUIDEmpty = errors.New("GUID cannot be empty")
	// ErrGUIDInvalid is returned when GUID format is invalid
	ErrGUIDInvalid = errors.New("GUID format is invalid")
)

// guidRegex validates UUID format (RFC 4122)
var guidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// GUID represents a globally unique identifier (UUID)
// It is immutable and ensures the GUID is always in a valid format
type GUID struct {
	value string
}

// NewGUID creates a new GUID value object from a string
// It validates the GUID format and normalizes it (lowercase)
// Returns an error if the GUID is invalid
func NewGUID(value string) (GUID, error) {
	if value == "" {
		return GUID{}, ErrGUIDEmpty
	}

	// Normalize to lowercase
	normalized := strings.ToLower(strings.TrimSpace(value))

	if !guidRegex.MatchString(normalized) {
		return GUID{}, ErrGUIDInvalid
	}

	return GUID{value: normalized}, nil
}

// NewGUIDFromString is an alias for NewGUID for clarity
func NewGUIDFromString(value string) (GUID, error) {
	return NewGUID(value)
}

// Value returns the GUID value as a string
func (g GUID) Value() string {
	return g.value
}

// String returns the string representation of the GUID
func (g GUID) String() string {
	return g.value
}

// Equals compares two GUID value objects for equality
func (g GUID) Equals(other GUID) bool {
	return g.value == other.value
}

// IsEmpty checks if the GUID is empty
func (g GUID) IsEmpty() bool {
	return g.value == ""
}

