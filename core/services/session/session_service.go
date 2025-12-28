package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := tracing.StartSpan(ctx, "session.create",
		trace.WithAttributes(attribute.String("session.user_id", userID)),
	)
	defer span.End()

	// Generate unique session ID
	sessionID, err := generateSessionID()
	if err != nil {
		tracing.RecordError(span, err)
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
		tracing.RecordError(span, err)
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	span.SetAttributes(
		attribute.String("session.id", sessionID),
		attribute.String("session.ttl", s.ttl.String()),
	)
	return sessionID, nil
}

// GetSession retrieves session data by sessionID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	ctx, span := tracing.StartSpan(ctx, "session.get",
		trace.WithAttributes(attribute.String("session.id", sessionID)),
	)
	defer span.End()

	data, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}
	return data, nil
}

// UpdateSession updates existing session data
func (s *SessionService) UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error {
	ctx, span := tracing.StartSpan(ctx, "session.update",
		trace.WithAttributes(attribute.String("session.id", sessionID)),
	)
	defer span.End()

	// Check if session exists
	exists, err := s.repo.Exists(ctx, sessionID)
	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if !exists {
		err := fmt.Errorf("session not found: %s", sessionID)
		tracing.RecordError(span, err)
		return err
	}

	// Get current session to preserve userID and createdAt
	currentData, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to get current session: %w", err)
	}

	// Merge data, preserving existing fields
	for k, v := range data {
		currentData[k] = v
	}

	// Store updated session with original TTL (will be refreshed)
	if err := s.repo.SetSession(ctx, sessionID, currentData, s.ttl); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession removes a session
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	ctx, span := tracing.StartSpan(ctx, "session.delete",
		trace.WithAttributes(attribute.String("session.id", sessionID)),
	)
	defer span.End()

	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// RefreshSession extends the TTL of an existing session
func (s *SessionService) RefreshSession(ctx context.Context, sessionID string) error {
	ctx, span := tracing.StartSpan(ctx, "session.refresh",
		trace.WithAttributes(
			attribute.String("session.id", sessionID),
			attribute.String("session.ttl", s.ttl.String()),
		),
	)
	defer span.End()

	if err := s.repo.RefreshSession(ctx, sessionID, s.ttl); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to refresh session: %w", err)
	}
	return nil
}

// SessionMetadata contains metadata about a session
type SessionMetadata struct {
	IPAddress   string
	UserAgent   string
	DeviceInfo  string
	LastActivity time.Time
}

// CreateSessionWithMetadata creates a new session with metadata (IP, user agent, etc.)
// Returns the generated sessionID
func (s *SessionService) CreateSessionWithMetadata(ctx context.Context, userID int64, metadata SessionMetadata) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "session.create_with_metadata",
		trace.WithAttributes(
			attribute.Int64("session.user_id", userID),
			attribute.String("session.ip_address", metadata.IPAddress),
		),
	)
	defer span.End()

	// Generate unique session ID
	sessionID, err := generateSessionID()
	if err != nil {
		tracing.RecordError(span, err)
		return "", err
	}

	// Build session data with metadata
	data := map[string]interface{}{
		"userID":       fmt.Sprintf("%d", userID),
		"userIDInt":    userID, // Store as int64 for easier retrieval
		"ipAddress":    metadata.IPAddress,
		"userAgent":    metadata.UserAgent,
		"deviceInfo":   metadata.DeviceInfo,
		"createdAt":    time.Now().Unix(),
		"lastActivity": metadata.LastActivity.Unix(),
	}

	// Store session
	if err := s.repo.SetSession(ctx, sessionID, data, s.ttl); err != nil {
		tracing.RecordError(span, err)
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Add session ID to user's session set
	if err := s.repo.AddUserSession(ctx, userID, sessionID); err != nil {
		tracing.RecordError(span, err)
		// Try to clean up the session if adding to set fails
		_ = s.repo.DeleteSession(ctx, sessionID)
		return "", fmt.Errorf("failed to add user session: %w", err)
	}

	span.SetAttributes(
		attribute.String("session.id", sessionID),
		attribute.String("session.ttl", s.ttl.String()),
	)
	return sessionID, nil
}

// GetUserSessions retrieves all sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	ctx, span := tracing.StartSpan(ctx, "session.get_user_sessions",
		trace.WithAttributes(attribute.Int64("session.user_id", userID)),
	)
	defer span.End()

	// Get all session IDs for the user
	sessionIDs, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Retrieve session data for each session ID
	sessions := make([]map[string]interface{}, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		sessionData, err := s.repo.GetSession(ctx, sessionID)
		if err != nil {
			// If session doesn't exist, remove it from the set (cleanup)
			_ = s.repo.RemoveUserSession(ctx, userID, sessionID)
			continue
		}
		// Add session ID to the data
		sessionData["id"] = sessionID
		sessions = append(sessions, sessionData)
	}

	return sessions, nil
}

// DeleteUserSession removes a specific session for a user
func (s *SessionService) DeleteUserSession(ctx context.Context, userID int64, sessionID string) error {
	ctx, span := tracing.StartSpan(ctx, "session.delete_user_session",
		trace.WithAttributes(
			attribute.Int64("session.user_id", userID),
			attribute.String("session.id", sessionID),
		),
	)
	defer span.End()

	// Verify session belongs to user
	sessionData, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if session belongs to user
	sessionUserID, ok := sessionData["userIDInt"].(int64)
	if !ok {
		// Fallback to string conversion
		userIDStr, ok := sessionData["userID"].(string)
		if !ok {
			return fmt.Errorf("invalid session data: userID not found")
		}
		var parseErr error
		if sessionUserID, parseErr = parseUserID(userIDStr); parseErr != nil {
			return fmt.Errorf("invalid session data: %w", parseErr)
		}
	}

	if sessionUserID != userID {
		return fmt.Errorf("session does not belong to user")
	}

	// Delete session
	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user's session set
	if err := s.repo.RemoveUserSession(ctx, userID, sessionID); err != nil {
		tracing.RecordError(span, err)
		// Don't fail if removal from set fails - session is already deleted
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a user
func (s *SessionService) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	ctx, span := tracing.StartSpan(ctx, "session.delete_all_user_sessions",
		trace.WithAttributes(attribute.Int64("session.user_id", userID)),
	)
	defer span.End()

	// Get all session IDs for the user
	sessionIDs, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete each session
	for _, sessionID := range sessionIDs {
		if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
			// Log error but continue deleting other sessions
			tracing.RecordError(span, err)
		}
		// Remove from set (best effort)
		_ = s.repo.RemoveUserSession(ctx, userID, sessionID)
	}

	return nil
}

// AssociateRefreshToken associates a refresh token hash with a session ID
func (s *SessionService) AssociateRefreshToken(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error {
	ctx, span := tracing.StartSpan(ctx, "session.associate_refresh_token",
		trace.WithAttributes(
			attribute.String("session.id", sessionID),
		),
	)
	defer span.End()

	if err := s.repo.SetRefreshTokenSession(ctx, refreshTokenHash, sessionID, ttl); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to associate refresh token with session: %w", err)
	}

	return nil
}

// GetSessionByRefreshToken retrieves session ID associated with a refresh token hash
func (s *SessionService) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "session.get_by_refresh_token")
	defer span.End()

	sessionID, err := s.repo.GetSessionByRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		tracing.RecordError(span, err)
		return "", err
	}

	return sessionID, nil
}

// DeleteRefreshTokenAssociation removes the association between refresh token and session
func (s *SessionService) DeleteRefreshTokenAssociation(ctx context.Context, refreshTokenHash string) error {
	ctx, span := tracing.StartSpan(ctx, "session.delete_refresh_token_association")
	defer span.End()

	if err := s.repo.DeleteRefreshTokenSession(ctx, refreshTokenHash); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to delete refresh token association: %w", err)
	}

	return nil
}

// parseUserID parses userID from string to int64
func parseUserID(userIDStr string) (int64, error) {
	var userID int64
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		return 0, fmt.Errorf("failed to parse userID: %w", err)
	}
	return userID, nil
}

