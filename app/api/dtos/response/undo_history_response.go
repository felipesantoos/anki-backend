package response

import "time"

// UndoHistoryResponse represents the response payload for an undo history record
type UndoHistoryResponse struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	OperationType string    `json:"operation_type"`
	OperationData string    `json:"operation_data"`
	CreatedAt     time.Time `json:"created_at"`
}

