package response

import "time"

// SessionResponse represents a session in the API response
// @Description Response payload for session information
type SessionResponse struct {
	// Session ID
	ID string `json:"id" example:"a1b2c3d4e5f6..."`

	// User ID
	UserID int64 `json:"user_id" example:"123"`

	// IP address from which the session was created
	IPAddress string `json:"ip_address" example:"192.168.1.1"`

	// User agent string
	UserAgent string `json:"user_agent" example:"Mozilla/5.0..."`

	// Device information (optional)
	DeviceInfo string `json:"device_info,omitempty" example:"Chrome on Windows"`

	// Session creation timestamp
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:00:00Z"`

	// Last activity timestamp
	LastActivity time.Time `json:"last_activity" example:"2024-01-15T10:30:00Z"`

	// Whether this is the current session
	IsCurrent bool `json:"is_current" example:"true"`
}

// SessionListResponse represents a list of sessions
// @Description Response payload for list of sessions
type SessionListResponse struct {
	// List of sessions
	Sessions []SessionResponse `json:"sessions"`

	// Total number of sessions
	Total int `json:"total" example:"3"`
}

