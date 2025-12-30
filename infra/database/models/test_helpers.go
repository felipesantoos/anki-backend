package models

import (
	"database/sql"
	"time"
)

// Helper functions for model tests

// sqlNullString creates a sql.NullString
func sqlNullString(s string, valid bool) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  valid,
	}
}

// sqlNullTime creates a sql.NullTime
func sqlNullTime(t time.Time, valid bool) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: valid,
	}
}

// sqlNullInt64 creates a sql.NullInt64
func sqlNullInt64(i int64, valid bool) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: valid,
	}
}

// sqlNullFloat64 creates a sql.NullFloat64
func sqlNullFloat64(f float64, valid bool) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   valid,
	}
}

