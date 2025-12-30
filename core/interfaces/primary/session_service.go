package primary

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/services/session"
)

// ISessionService defines the interface for session management
type ISessionService interface {
	CreateSessionWithMetadata(ctx context.Context, userID int64, metadata session.SessionMetadata) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserSessions(ctx context.Context, userID int64) ([]map[string]interface{}, error)
	DeleteUserSession(ctx context.Context, userID int64, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID int64) error
	AssociateRefreshToken(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error
	GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error)
	DeleteRefreshTokenAssociation(ctx context.Context, refreshTokenHash string) error
	UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error
}

