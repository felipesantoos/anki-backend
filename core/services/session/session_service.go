package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SessionService provides high-level session management operations
// Uses session repository interface (Redis, database, etc.)
type SessionService struct {
	repo    secondary.ISessionRepository
	ttl     time.Duration
}

// NewSessionService creates a new SessionService instance
func NewSessionService(repo secondary.ISessionRepository, defaultTTL time.Duration) *SessionService {
	return &SessionService{
		repo: repo,
		ttl:  defaultTTL,
	}
}

// generateSessionID generates a secure random session ID
// Uses 32 random bytes (256 bits) encoded as hex (64 characters)
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session with the given userID and data
// Returns the generated sessionID
func (s *SessionService) CreateSession(ctx context.Context, userID string, data map[string]interface{}) (string, error) {
	// Generate unique session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	// Add userID to session data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["userID"] = userID
	data["createdAt"] = time.Now().Unix()

	// Store session
	if err := s.repo.SetSession(ctx, sessionID, data, s.ttl); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// GetSession retrieves session data by sessionID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	return s.repo.GetSession(ctx, sessionID)
}

// UpdateSession updates existing session data
func (s *SessionService) UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error {
	// Check if session exists
	exists, err := s.repo.Exists(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Get current session to preserve userID and createdAt
	currentData, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get current session: %w", err)
	}

	// Merge data, preserving existing fields
	for k, v := range data {
		currentData[k] = v
	}

	// Store updated session with original TTL (will be refreshed)
	if err := s.repo.SetSession(ctx, sessionID, currentData, s.ttl); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession removes a session
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// RefreshSession extends the TTL of an existing session
func (s *SessionService) RefreshSession(ctx context.Context, sessionID string) error {
	if err := s.repo.RefreshSession(ctx, sessionID, s.ttl); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}
	return nil
}

