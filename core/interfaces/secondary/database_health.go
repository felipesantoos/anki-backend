package secondary

import (
	"context"
	"database/sql"
)

// IDatabaseRepository defines the interface for database health checks
// Implementation agnostic - works with PostgreSQL, MySQL, etc.
type IDatabaseRepository interface {
	Ping(ctx context.Context) error
	GetDB() *sql.DB
}

