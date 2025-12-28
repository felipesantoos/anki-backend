package entities

import (
	"path/filepath"
	"strings"
	"time"
)

// Media represents a media file entity in the domain
// Media files include images, audio, and video used in notes
type Media struct {
	id          int64
	userID      int64
	filename    string
	hash        string // SHA-256 hash for deduplication
	size        int64
	mimeType    string
	storagePath string
	createdAt   time.Time
	deletedAt   *time.Time
}

// Getters
func (m *Media) GetID() int64 {
	return m.id
}

func (m *Media) GetUserID() int64 {
	return m.userID
}

func (m *Media) GetFilename() string {
	return m.filename
}

func (m *Media) GetHash() string {
	return m.hash
}

func (m *Media) GetSize() int64 {
	return m.size
}

func (m *Media) GetMimeType() string {
	return m.mimeType
}

func (m *Media) GetStoragePath() string {
	return m.storagePath
}

func (m *Media) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m *Media) GetDeletedAt() *time.Time {
	return m.deletedAt
}

// Setters
func (m *Media) SetID(id int64) {
	m.id = id
}

func (m *Media) SetUserID(userID int64) {
	m.userID = userID
}

func (m *Media) SetFilename(filename string) {
	m.filename = filename
}

func (m *Media) SetHash(hash string) {
	m.hash = hash
}

func (m *Media) SetSize(size int64) {
	m.size = size
}

func (m *Media) SetMimeType(mimeType string) {
	m.mimeType = mimeType
}

func (m *Media) SetStoragePath(storagePath string) {
	m.storagePath = storagePath
}

func (m *Media) SetCreatedAt(createdAt time.Time) {
	m.createdAt = createdAt
}

func (m *Media) SetDeletedAt(deletedAt *time.Time) {
	m.deletedAt = deletedAt
}

// IsActive checks if the media file is active (not deleted)
func (m *Media) IsActive() bool {
	return m.deletedAt == nil
}

// GetFileExtension returns the file extension from the filename
func (m *Media) GetFileExtension() string {
	ext := filepath.Ext(m.filename)
	return strings.ToLower(ext)
}

