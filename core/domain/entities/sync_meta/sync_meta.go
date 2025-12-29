package syncmeta

import (
	"time"
)

// SyncMeta represents synchronization metadata entity in the domain
// It tracks sync state for each user and client device
type SyncMeta struct {
	id          int64
	userID      int64
	clientID    string
	lastSync    time.Time
	lastSyncUSN int64 // Update Sequence Number
	createdAt   time.Time
	updatedAt   time.Time
}

// Getters
func (sm *SyncMeta) GetID() int64 {
	return sm.id
}

func (sm *SyncMeta) GetUserID() int64 {
	return sm.userID
}

func (sm *SyncMeta) GetClientID() string {
	return sm.clientID
}

func (sm *SyncMeta) GetLastSync() time.Time {
	return sm.lastSync
}

func (sm *SyncMeta) GetLastSyncUSN() int64 {
	return sm.lastSyncUSN
}

func (sm *SyncMeta) GetCreatedAt() time.Time {
	return sm.createdAt
}

func (sm *SyncMeta) GetUpdatedAt() time.Time {
	return sm.updatedAt
}

// Setters
func (sm *SyncMeta) SetID(id int64) {
	sm.id = id
}

func (sm *SyncMeta) SetUserID(userID int64) {
	sm.userID = userID
}

func (sm *SyncMeta) SetClientID(clientID string) {
	sm.clientID = clientID
}

func (sm *SyncMeta) SetLastSync(lastSync time.Time) {
	sm.lastSync = lastSync
}

func (sm *SyncMeta) SetLastSyncUSN(lastSyncUSN int64) {
	sm.lastSyncUSN = lastSyncUSN
}

func (sm *SyncMeta) SetCreatedAt(createdAt time.Time) {
	sm.createdAt = createdAt
}

func (sm *SyncMeta) SetUpdatedAt(updatedAt time.Time) {
	sm.updatedAt = updatedAt
}

