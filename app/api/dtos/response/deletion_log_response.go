package response

import (
	"time"
)

// DeletionLogResponse represents the response payload for a deletion log
type DeletionLogResponse struct {
	ID         int64                  `json:"id"`
	UserID     int64                  `json:"user_id"`
	ObjectType string                 `json:"object_type"`
	ObjectID   int64                  `json:"object_id"`
	ObjectData map[string]interface{} `json:"object_data,omitempty"`
	DeletedAt  time.Time              `json:"deleted_at"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// RecentDeletionsResponse represents the response payload for recent deletions endpoint
type RecentDeletionsResponse struct {
	Data       []*DeletionLogResponse `json:"data"`
	Pagination PaginationResponse     `json:"pagination"`
}

