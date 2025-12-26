package secondary

import "context"

// IDatabaseRepository defines the interface for database health checks
// Implementation agnostic - works with PostgreSQL, MySQL, etc.
type IDatabaseRepository interface {
	Ping(ctx context.Context) error
}

