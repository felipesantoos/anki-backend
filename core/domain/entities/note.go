package entities

import (
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// Note represents a note entity in the domain
// A note is the source data that generates one or more cards
type Note struct {
	ID          int64
	UserID      int64
	GUID        valueobjects.GUID
	NoteTypeID  int64
	FieldsJSON  string   // Object JSON in database
	Tags        []string // Array in database
	Marked      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// IsActive checks if the note is active (not deleted)
func (n *Note) IsActive() bool {
	return n.DeletedAt == nil
}

// GetFirstField returns the first field value from FieldsJSON
// This is used for duplicate detection
// Note: This is a simplified implementation - in production, use the field order from NoteType
func (n *Note) GetFirstField() string {
	// Simple extraction - in production, parse JSON properly and use NoteType field order
	// For now, return empty string as parsing should be done in service layer
	return ""
}

// HasTag checks if the note has a specific tag
func (n *Note) HasTag(tag string) bool {
	if tag == "" {
		return false
	}

	tagLower := strings.ToLower(strings.TrimSpace(tag))
	for _, t := range n.Tags {
		if strings.ToLower(strings.TrimSpace(t)) == tagLower {
			return true
		}
	}
	return false
}

// AddTag adds a tag to the note if it doesn't already exist
func (n *Note) AddTag(tag string) {
	if tag == "" {
		return
	}

	tagTrimmed := strings.TrimSpace(tag)
	if n.HasTag(tagTrimmed) {
		return
	}

	n.Tags = append(n.Tags, tagTrimmed)
	n.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the note
func (n *Note) RemoveTag(tag string) {
	if tag == "" {
		return
	}

	tagLower := strings.ToLower(strings.TrimSpace(tag))
	newTags := make([]string, 0, len(n.Tags))
	for _, t := range n.Tags {
		if strings.ToLower(strings.TrimSpace(t)) != tagLower {
			newTags = append(newTags, t)
		}
	}
	n.Tags = newTags
	n.UpdatedAt = time.Now()
}

// IsMarked checks if the note is marked
func (n *Note) IsMarked() bool {
	return n.Marked
}

// Mark marks the note
func (n *Note) Mark() {
	if !n.Marked {
		n.Marked = true
		n.UpdatedAt = time.Now()
	}
}

// Unmark unmarks the note
func (n *Note) Unmark() {
	if n.Marked {
		n.Marked = false
		n.UpdatedAt = time.Now()
	}
}

