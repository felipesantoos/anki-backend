package entities

import (
	"path/filepath"
	"strings"
	"time"
)

// Media represents a media file entity in the domain
// Media files include images, audio, and video used in notes
type Media struct {
	ID          int64
	UserID      int64
	Filename    string
	Hash        string // SHA-256 hash for deduplication
	Size        int64
	MimeType    string
	StoragePath string
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

// IsActive checks if the media file is active (not deleted)
func (m *Media) IsActive() bool {
	return m.DeletedAt == nil
}

// GetFileExtension returns the file extension from the filename
func (m *Media) GetFileExtension() string {
	ext := filepath.Ext(m.Filename)
	return strings.ToLower(ext)
}

