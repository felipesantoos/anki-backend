package response

import "time"

// FilteredDeckResponse represents the response payload for a filtered deck
// @Description Response payload containing filtered deck information
type FilteredDeckResponse struct {
	// Unique identifier for the filtered deck
	ID int64 `json:"id" example:"1"`

	// ID of the user who owns the deck
	UserID int64 `json:"user_id" example:"1"`

	// Name of the filtered deck
	Name string `json:"name" example:"Revisão de Hoje"`

	// Search filter criteria
	SearchFilter string `json:"search_filter" example:"is:due"`

	// Maximum number of cards
	Limit int `json:"limit" example:"100"`

	// Order by criteria
	OrderBy string `json:"order_by" example:"random"`

	// Whether rescheduling is enabled
	Reschedule bool `json:"reschedule" example:"true"`

	// Timestamp when created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp when last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

