package note

import (
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// Note represents a note entity in the domain
// A note is the source data that generates one or more cards
type Note struct {
	id          int64
	userID      int64
	guid        valueobjects.GUID
	noteTypeID  int64
	fieldsJSON  string   // Object JSON in database
	tags        []string // Array in database
	marked      bool
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// Getters
func (n *Note) GetID() int64 {
	return n.id
}

func (n *Note) GetUserID() int64 {
	return n.userID
}

func (n *Note) GetGUID() valueobjects.GUID {
	return n.guid
}

func (n *Note) GetNoteTypeID() int64 {
	return n.noteTypeID
}

func (n *Note) GetFieldsJSON() string {
	return n.fieldsJSON
}

func (n *Note) GetTags() []string {
	return n.tags
}

func (n *Note) GetMarked() bool {
	return n.marked
}

func (n *Note) GetCreatedAt() time.Time {
	return n.createdAt
}

func (n *Note) GetUpdatedAt() time.Time {
	return n.updatedAt
}

func (n *Note) GetDeletedAt() *time.Time {
	return n.deletedAt
}

// Setters
func (n *Note) SetID(id int64) {
	n.id = id
}

func (n *Note) SetUserID(userID int64) {
	n.userID = userID
}

func (n *Note) SetGUID(guid valueobjects.GUID) {
	n.guid = guid
}

func (n *Note) SetNoteTypeID(noteTypeID int64) {
	n.noteTypeID = noteTypeID
}

func (n *Note) SetFieldsJSON(fieldsJSON string) {
	n.fieldsJSON = fieldsJSON
}

func (n *Note) SetTags(tags []string) {
	n.tags = tags
}

func (n *Note) SetMarked(marked bool) {
	n.marked = marked
}

func (n *Note) SetCreatedAt(createdAt time.Time) {
	n.createdAt = createdAt
}

func (n *Note) SetUpdatedAt(updatedAt time.Time) {
	n.updatedAt = updatedAt
}

func (n *Note) SetDeletedAt(deletedAt *time.Time) {
	n.deletedAt = deletedAt
}

// IsActive checks if the note is active (not deleted)
func (n *Note) IsActive() bool {
	return n.deletedAt == nil
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
	for _, t := range n.tags {
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

	n.tags = append(n.tags, tagTrimmed)
	n.updatedAt = time.Now()
}

// RemoveTag removes a tag from the note
func (n *Note) RemoveTag(tag string) {
	if tag == "" {
		return
	}

	tagLower := strings.ToLower(strings.TrimSpace(tag))
	newTags := make([]string, 0, len(n.tags))
	for _, t := range n.tags {
		if strings.ToLower(strings.TrimSpace(t)) != tagLower {
			newTags = append(newTags, t)
		}
	}
	n.tags = newTags
	n.updatedAt = time.Now()
}

// IsMarked checks if the note is marked
func (n *Note) IsMarked() bool {
	return n.marked
}

// Mark marks the note
func (n *Note) Mark() {
	if !n.marked {
		n.marked = true
		n.updatedAt = time.Now()
	}
}

// Unmark unmarks the note
func (n *Note) Unmark() {
	if n.marked {
		n.marked = false
		n.updatedAt = time.Now()
	}
}

// NoteFilters represents the filters for listing notes
type NoteFilters struct {
	DeckID     *int64
	NoteTypeID *int64
	Tags       []string
	Search     string
	Limit      int
	Offset     int
}

// DuplicateGroup represents a group of duplicate notes with the same field value
type DuplicateGroup struct {
	FieldValue string
	Notes      []*DuplicateNoteInfo
}

// DuplicateNoteInfo contains basic information about a duplicate note
type DuplicateNoteInfo struct {
	ID        int64
	GUID      string
	DeckID    int64
	CreatedAt time.Time
}

// DuplicateResult contains the result of a duplicate search
type DuplicateResult struct {
	Duplicates []*DuplicateGroup
	Total      int
}

