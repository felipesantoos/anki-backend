package entities

import (
	"time"
)

// Helper functions for tests
func timePtr(t time.Time) *time.Time {
	return &t
}

func int64Ptr(i int64) *int64 {
	return &i
}

