package response

import "time"

// DeletionLogResponse represents the response payload for a deletion log
type DeletionLogResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	ObjectType string    `json:"object_type"`
	ObjectID   int64     `json:"object_id"`
	DeletedAt  time.Time `json:"deleted_at"`
}

