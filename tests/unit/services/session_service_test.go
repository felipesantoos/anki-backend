package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/services/session"
)

// mockSessionRepository is a mock implementation of ISessionRepository for testing
type mockSessionRepository struct {
	getSessionFunc              func(ctx context.Context, sessionID string) (map[string]interface{}, error)
	setSessionFunc              func(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error
	deleteSessionFunc           func(ctx context.Context, sessionID string) error
	refreshFunc                 func(ctx context.Context, sessionID string, ttl time.Duration) error
	existsFunc                  func(ctx context.Context, sessionID string) (bool, error)
	getUserSessionsFunc         func(ctx context.Context, userID int64) ([]string, error)
	addUserSessionFunc          func(ctx context.Context, userID int64, sessionID string) error
	removeUserSessionFunc       func(ctx context.Context, userID int64, sessionID string) error
	getSessionByRefreshTokenFunc func(ctx context.Context, refreshTokenHash string) (string, error)
	setRefreshTokenSessionFunc  func(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error
	deleteRefreshTokenSessionFunc func(ctx context.Context, refreshTokenHash string) error
}

func (m *mockSessionRepository) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	if m.getSessionFunc != nil {
		return m.getSessionFunc(ctx, sessionID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepository) SetSession(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error {
	if m.setSessionFunc != nil {
		return m.setSessionFunc(ctx, sessionID, data, ttl)
	}
	return errors.New("not implemented")
}

func (m *mockSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, sessionID)
	}
	return errors.New("not implemented")
}

func (m *mockSessionRepository) RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	if m.refreshFunc != nil {
		return m.refreshFunc(ctx, sessionID, ttl)
	}
	return errors.New("not implemented")
}

func (m *mockSessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, sessionID)
	}
	return false, errors.New("not implemented")
}

func (m *mockSessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]string, error) {
	if m.getUserSessionsFunc != nil {
		return m.getUserSessionsFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepository) AddUserSession(ctx context.Context, userID int64, sessionID string) error {
	if m.addUserSessionFunc != nil {
		return m.addUserSessionFunc(ctx, userID, sessionID)
	}
	return errors.New("not implemented")
}

func (m *mockSessionRepository) RemoveUserSession(ctx context.Context, userID int64, sessionID string) error {
	if m.removeUserSessionFunc != nil {
		return m.removeUserSessionFunc(ctx, userID, sessionID)
	}
	return errors.New("not implemented")
}

func (m *mockSessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error) {
	if m.getSessionByRefreshTokenFunc != nil {
		return m.getSessionByRefreshTokenFunc(ctx, refreshTokenHash)
	}
	return "", errors.New("not implemented")
}

func (m *mockSessionRepository) SetRefreshTokenSession(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error {
	if m.setRefreshTokenSessionFunc != nil {
		return m.setRefreshTokenSessionFunc(ctx, refreshTokenHash, sessionID, ttl)
	}
	return errors.New("not implemented")
}

func (m *mockSessionRepository) DeleteRefreshTokenSession(ctx context.Context, refreshTokenHash string) error {
	if m.deleteRefreshTokenSessionFunc != nil {
		return m.deleteRefreshTokenSessionFunc(ctx, refreshTokenHash)
	}
	return errors.New("not implemented")
}

func TestSessionService_CreateSession(t *testing.T) {
	var capturedSessionID string
	var capturedData map[string]interface{}
	var capturedTTL time.Duration

	repo := &mockSessionRepository{
		setSessionFunc: func(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error {
			capturedSessionID = sessionID
			capturedData = data
			capturedTTL = ttl
			return nil
		},
	}

	ttl := 30 * time.Minute
	service := session.NewSessionService(repo, ttl)

	userID := "user123"
	data := map[string]interface{}{
		"email": "user@example.com",
	}

	sessionID, err := service.CreateSession(context.Background(), userID, data)
	if err != nil {
		t.Fatalf("CreateSession() error = %v, want nil", err)
	}

	if sessionID == "" {
		t.Error("CreateSession() sessionID is empty")
	}

	if len(sessionID) < 32 {
		t.Errorf("CreateSession() sessionID length = %d, want at least 32", len(sessionID))
	}

	// Check that userID was added to data
	if capturedData["userID"] != userID {
		t.Errorf("CreateSession() data[userID] = %v, want %v", capturedData["userID"], userID)
	}

	// Check that createdAt was added
	if capturedData["createdAt"] == nil {
		t.Error("CreateSession() data[createdAt] is nil")
	}

	// Check that original data was preserved
	if capturedData["email"] != "user@example.com" {
		t.Errorf("CreateSession() data[email] = %v, want user@example.com", capturedData["email"])
	}

	if capturedTTL != ttl {
		t.Errorf("CreateSession() TTL = %v, want %v", capturedTTL, ttl)
	}

	if capturedSessionID != sessionID {
		t.Errorf("CreateSession() sessionID = %v, want %v", capturedSessionID, sessionID)
	}
}

func TestSessionService_GetSession(t *testing.T) {
	expectedData := map[string]interface{}{
		"userID":    "user123",
		"email":     "user@example.com",
		"createdAt": int64(1234567890),
	}

	repo := &mockSessionRepository{
		getSessionFunc: func(ctx context.Context, sessionID string) (map[string]interface{}, error) {
			return expectedData, nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	data, err := service.GetSession(context.Background(), "test-session-id")
	if err != nil {
		t.Fatalf("GetSession() error = %v, want nil", err)
	}

	if data["userID"] != expectedData["userID"] {
		t.Errorf("GetSession() data[userID] = %v, want %v", data["userID"], expectedData["userID"])
	}
}

func TestSessionService_UpdateSession(t *testing.T) {
	existingData := map[string]interface{}{
		"userID":    "user123",
		"email":     "old@example.com",
		"createdAt": int64(1234567890),
	}

	var updatedData map[string]interface{}

	repo := &mockSessionRepository{
		existsFunc: func(ctx context.Context, sessionID string) (bool, error) {
			return true, nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (map[string]interface{}, error) {
			return existingData, nil
		},
		setSessionFunc: func(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error {
			updatedData = data
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	updateData := map[string]interface{}{
		"email": "new@example.com",
	}

	err := service.UpdateSession(context.Background(), "test-session-id", updateData)
	if err != nil {
		t.Fatalf("UpdateSession() error = %v, want nil", err)
	}

	// Check that email was updated
	if updatedData["email"] != "new@example.com" {
		t.Errorf("UpdateSession() updatedData[email] = %v, want new@example.com", updatedData["email"])
	}

	// Check that userID was preserved
	if updatedData["userID"] != "user123" {
		t.Errorf("UpdateSession() updatedData[userID] = %v, want user123", updatedData["userID"])
	}

	// Check that createdAt was preserved
	if updatedData["createdAt"] != int64(1234567890) {
		t.Errorf("UpdateSession() updatedData[createdAt] = %v, want 1234567890", updatedData["createdAt"])
	}
}

func TestSessionService_DeleteSession(t *testing.T) {
	var capturedSessionID string

	repo := &mockSessionRepository{
		deleteSessionFunc: func(ctx context.Context, sessionID string) error {
			capturedSessionID = sessionID
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	sessionID := "test-session-id"
	err := service.DeleteSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("DeleteSession() error = %v, want nil", err)
	}

	if capturedSessionID != sessionID {
		t.Errorf("DeleteSession() captured sessionID = %v, want %v", capturedSessionID, sessionID)
	}
}

func TestSessionService_RefreshSession(t *testing.T) {
	var capturedSessionID string
	var capturedTTL time.Duration

	repo := &mockSessionRepository{
		refreshFunc: func(ctx context.Context, sessionID string, ttl time.Duration) error {
			capturedSessionID = sessionID
			capturedTTL = ttl
			return nil
		},
	}

	ttl := 30 * time.Minute
	service := session.NewSessionService(repo, ttl)

	sessionID := "test-session-id"
	err := service.RefreshSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("RefreshSession() error = %v, want nil", err)
	}

	if capturedSessionID != sessionID {
		t.Errorf("RefreshSession() captured sessionID = %v, want %v", capturedSessionID, sessionID)
	}

	if capturedTTL != ttl {
		t.Errorf("RefreshSession() captured TTL = %v, want %v", capturedTTL, ttl)
	}
}

func TestSessionService_CreateSessionWithMetadata(t *testing.T) {
	var capturedData map[string]interface{}
	var capturedUserID int64
	var capturedSessionIDForSet string

	repo := &mockSessionRepository{
		setSessionFunc: func(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error {
			capturedData = data
			return nil
		},
		addUserSessionFunc: func(ctx context.Context, userID int64, sessionID string) error {
			capturedUserID = userID
			capturedSessionIDForSet = sessionID
			return nil
		},
	}

	ttl := 30 * time.Minute
	service := session.NewSessionService(repo, ttl)

	userID := int64(123)
	metadata := session.SessionMetadata{
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		DeviceInfo:   "Chrome on Windows",
		LastActivity: time.Now(),
	}

	sessionID, err := service.CreateSessionWithMetadata(context.Background(), userID, metadata)
	if err != nil {
		t.Fatalf("CreateSessionWithMetadata() error = %v, want nil", err)
	}

	if sessionID == "" {
		t.Error("CreateSessionWithMetadata() sessionID is empty")
	}

	if capturedUserID != userID {
		t.Errorf("CreateSessionWithMetadata() capturedUserID = %v, want %v", capturedUserID, userID)
	}

	if capturedSessionIDForSet != sessionID {
		t.Errorf("CreateSessionWithMetadata() capturedSessionIDForSet = %v, want %v", capturedSessionIDForSet, sessionID)
	}

	// Check metadata was stored
	if capturedData["ipAddress"] != metadata.IPAddress {
		t.Errorf("CreateSessionWithMetadata() data[ipAddress] = %v, want %v", capturedData["ipAddress"], metadata.IPAddress)
	}

	if capturedData["userAgent"] != metadata.UserAgent {
		t.Errorf("CreateSessionWithMetadata() data[userAgent] = %v, want %v", capturedData["userAgent"], metadata.UserAgent)
	}
}

func TestSessionService_GetUserSessions(t *testing.T) {
	userID := int64(123)
	sessionIDs := []string{"session1", "session2", "session3"}
	sessionData1 := map[string]interface{}{
		"userID":    "123",
		"userIDInt": int64(123),
		"ipAddress": "192.168.1.1",
	}
	sessionData2 := map[string]interface{}{
		"userID":    "123",
		"userIDInt": int64(123),
		"ipAddress": "192.168.1.2",
	}

	callCount := 0
	repo := &mockSessionRepository{
		getUserSessionsFunc: func(ctx context.Context, uid int64) ([]string, error) {
			if uid != userID {
				t.Errorf("GetUserSessions() userID = %v, want %v", uid, userID)
			}
			return sessionIDs, nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (map[string]interface{}, error) {
			callCount++
			if sessionID == "session1" {
				return sessionData1, nil
			}
			if sessionID == "session2" {
				return sessionData2, nil
			}
			return nil, errors.New("session not found")
		},
		removeUserSessionFunc: func(ctx context.Context, uid int64, sid string) error {
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	sessions, err := service.GetUserSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserSessions() error = %v, want nil", err)
	}

	// Should have 2 valid sessions (session3 doesn't exist, so it's removed)
	if len(sessions) != 2 {
		t.Errorf("GetUserSessions() len(sessions) = %v, want 2", len(sessions))
	}

	// Check that IDs were added
	if sessions[0]["id"] == nil {
		t.Error("GetUserSessions() session[0][id] is nil")
	}
}

func TestSessionService_DeleteUserSession(t *testing.T) {
	userID := int64(123)
	sessionID := "test-session-id"
	sessionData := map[string]interface{}{
		"userID":    "123",
		"userIDInt": int64(123),
	}

	var deletedSessionID string
	var removedUserID int64
	var removedSessionID string

	repo := &mockSessionRepository{
		getSessionFunc: func(ctx context.Context, sid string) (map[string]interface{}, error) {
			if sid != sessionID {
				t.Errorf("GetSession() sessionID = %v, want %v", sid, sessionID)
			}
			return sessionData, nil
		},
		deleteSessionFunc: func(ctx context.Context, sid string) error {
			deletedSessionID = sid
			return nil
		},
		removeUserSessionFunc: func(ctx context.Context, uid int64, sid string) error {
			removedUserID = uid
			removedSessionID = sid
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	err := service.DeleteUserSession(context.Background(), userID, sessionID)
	if err != nil {
		t.Fatalf("DeleteUserSession() error = %v, want nil", err)
	}

	if deletedSessionID != sessionID {
		t.Errorf("DeleteUserSession() deletedSessionID = %v, want %v", deletedSessionID, sessionID)
	}

	if removedUserID != userID {
		t.Errorf("DeleteUserSession() removedUserID = %v, want %v", removedUserID, userID)
	}

	if removedSessionID != sessionID {
		t.Errorf("DeleteUserSession() removedSessionID = %v, want %v", removedSessionID, sessionID)
	}
}

func TestSessionService_DeleteUserSession_WrongUser(t *testing.T) {
	userID := int64(123)
	sessionID := "test-session-id"
	sessionData := map[string]interface{}{
		"userID":    "456",
		"userIDInt": int64(456),
	}

	repo := &mockSessionRepository{
		getSessionFunc: func(ctx context.Context, sid string) (map[string]interface{}, error) {
			return sessionData, nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	err := service.DeleteUserSession(context.Background(), userID, sessionID)
	if err == nil {
		t.Error("DeleteUserSession() error = nil, want error")
	}

	if err.Error() != "session does not belong to user" {
		t.Errorf("DeleteUserSession() error = %v, want 'session does not belong to user'", err)
	}
}

func TestSessionService_DeleteAllUserSessions(t *testing.T) {
	userID := int64(123)
	sessionIDs := []string{"session1", "session2"}

	var deletedSessions []string

	repo := &mockSessionRepository{
		getUserSessionsFunc: func(ctx context.Context, uid int64) ([]string, error) {
			if uid != userID {
				t.Errorf("GetUserSessions() userID = %v, want %v", uid, userID)
			}
			return sessionIDs, nil
		},
		deleteSessionFunc: func(ctx context.Context, sid string) error {
			deletedSessions = append(deletedSessions, sid)
			return nil
		},
		removeUserSessionFunc: func(ctx context.Context, uid int64, sid string) error {
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	err := service.DeleteAllUserSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("DeleteAllUserSessions() error = %v, want nil", err)
	}

	if len(deletedSessions) != 2 {
		t.Errorf("DeleteAllUserSessions() len(deletedSessions) = %v, want 2", len(deletedSessions))
	}
}

func TestSessionService_AssociateRefreshToken(t *testing.T) {
	refreshTokenHash := "hash123"
	sessionID := "session123"
	ttl := 30 * time.Minute

	var capturedHash string
	var capturedSessionID string
	var capturedTTL time.Duration

	repo := &mockSessionRepository{
		setRefreshTokenSessionFunc: func(ctx context.Context, hash string, sid string, t time.Duration) error {
			capturedHash = hash
			capturedSessionID = sid
			capturedTTL = t
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	err := service.AssociateRefreshToken(context.Background(), refreshTokenHash, sessionID, ttl)
	if err != nil {
		t.Fatalf("AssociateRefreshToken() error = %v, want nil", err)
	}

	if capturedHash != refreshTokenHash {
		t.Errorf("AssociateRefreshToken() capturedHash = %v, want %v", capturedHash, refreshTokenHash)
	}

	if capturedSessionID != sessionID {
		t.Errorf("AssociateRefreshToken() capturedSessionID = %v, want %v", capturedSessionID, sessionID)
	}

	if capturedTTL != ttl {
		t.Errorf("AssociateRefreshToken() capturedTTL = %v, want %v", capturedTTL, ttl)
	}
}

func TestSessionService_GetSessionByRefreshToken(t *testing.T) {
	refreshTokenHash := "hash123"
	expectedSessionID := "session123"

	repo := &mockSessionRepository{
		getSessionByRefreshTokenFunc: func(ctx context.Context, hash string) (string, error) {
			if hash != refreshTokenHash {
				t.Errorf("GetSessionByRefreshToken() hash = %v, want %v", hash, refreshTokenHash)
			}
			return expectedSessionID, nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	sessionID, err := service.GetSessionByRefreshToken(context.Background(), refreshTokenHash)
	if err != nil {
		t.Fatalf("GetSessionByRefreshToken() error = %v, want nil", err)
	}

	if sessionID != expectedSessionID {
		t.Errorf("GetSessionByRefreshToken() sessionID = %v, want %v", sessionID, expectedSessionID)
	}
}

func TestSessionService_DeleteRefreshTokenAssociation(t *testing.T) {
	refreshTokenHash := "hash123"

	var capturedHash string

	repo := &mockSessionRepository{
		deleteRefreshTokenSessionFunc: func(ctx context.Context, hash string) error {
			capturedHash = hash
			return nil
		},
	}

	service := session.NewSessionService(repo, time.Hour)

	err := service.DeleteRefreshTokenAssociation(context.Background(), refreshTokenHash)
	if err != nil {
		t.Fatalf("DeleteRefreshTokenAssociation() error = %v, want nil", err)
	}

	if capturedHash != refreshTokenHash {
		t.Errorf("DeleteRefreshTokenAssociation() capturedHash = %v, want %v", capturedHash, refreshTokenHash)
	}
}

