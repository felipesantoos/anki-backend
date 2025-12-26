package response

import "time"

// ErrorResponse represents a standardized error response
// @Description Standardized error response format returned by the API
type ErrorResponse struct {
	Error     string `json:"error" example:"NOT_FOUND"`                    // Error code (e.g., "NOT_FOUND", "VALIDATION_ERROR")
	Message   string `json:"message" example:"Resource not found"`         // Human-readable error message
	Code      string `json:"code" example:"NOT_FOUND"`                     // Error code (same as error field, kept for compatibility)
	RequestID string `json:"request_id" example:"abc123def456"`            // Request ID for tracking
	Timestamp string `json:"timestamp" example:"2024-01-15T10:30:00Z"`     // ISO 8601 timestamp
	Path      string `json:"path" example:"/api/users/123"`                // Request path
}

// NewErrorResponse creates a new ErrorResponse with the given error code and message
func NewErrorResponse(errorCode, message, requestID, path string) *ErrorResponse {
	return &ErrorResponse{
		Error:     errorCode,
		Message:   message,
		Code:      errorCode,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      path,
	}
}
