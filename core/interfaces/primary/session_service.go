package primary

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/services/session"
)

// ISessionService defines the interface for session management
type ISessionService interface {
	CreateSession(ctx context.Context, userID string, data map[string]interface{}) (string, error)
	CreateSessionWithMetadata(ctx context.Context, userID int64, metadata session.SessionMetadata) (string, error)
	GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error)
	DeleteSession(ctx context.Context, sessionID string) error
	RefreshSession(ctx context.Context, sessionID string) error
	GetUserSessions(ctx context.Context, userID int64) ([]map[string]interface{}, error)
	DeleteUserSession(ctx context.Context, userID int64, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID int64) error
	AssociateRefreshToken(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error
	GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error)
	DeleteRefreshTokenAssociation(ctx context.Context, refreshTokenHash string) error
	UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error
}

