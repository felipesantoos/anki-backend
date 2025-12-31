package secondary

import "context"

// ITransactionManager defines the interface for database transaction management
type ITransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

