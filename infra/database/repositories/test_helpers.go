package repositories

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// setupMockDB creates a mock database connection for unit tests
func setupMockDB(t require.TestingT) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	cleanup := func() {
		db.Close()
	}

	return db, mock, cleanup
}

