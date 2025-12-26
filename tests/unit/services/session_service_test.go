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
	getSessionFunc    func(ctx context.Context, sessionID string) (map[string]interface{}, error)
	setSessionFunc    func(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error
	deleteSessionFunc func(ctx context.Context, sessionID string) error
	refreshFunc       func(ctx context.Context, sessionID string, ttl time.Duration) error
	existsFunc        func(ctx context.Context, sessionID string) (bool, error)
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

