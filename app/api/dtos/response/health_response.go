package response

import "time"

// ComponentHealth represents the health status of a single component
// @Description Health status of a single component (database, redis, etc.)
type ComponentHealth struct {
	Status  string `json:"status" example:"healthy"`   // "healthy" or "unhealthy"
	Message string `json:"message" example:"Connection successful"` // Status message or error description
}

// HealthResponse represents the overall health check response
// @Description Health check response containing overall status, timestamp, and component health statuses
type HealthResponse struct {
	Status     string                      `json:"status" example:"healthy"`     // "healthy", "degraded", or "unhealthy"
	Timestamp  string                      `json:"timestamp" example:"2024-01-15T10:30:00Z"` // ISO 8601 timestamp
	Components map[string]ComponentHealth  `json:"components"`                  // Component health statuses
}

// NewHealthResponse creates a new HealthResponse
func NewHealthResponse() *HealthResponse {
	return &HealthResponse{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Components: make(map[string]ComponentHealth),
	}
}

// SetComponent sets the health status for a component
func (h *HealthResponse) SetComponent(name string, status string, message string) {
	h.Components[name] = ComponentHealth{
		Status:  status,
		Message: message,
	}
}

// CalculateOverallStatus calculates the overall status based on component statuses
func (h *HealthResponse) CalculateOverallStatus() {
	allHealthy := true
	anyHealthy := false

	for _, component := range h.Components {
		if component.Status == "healthy" {
			anyHealthy = true
		} else {
			allHealthy = false
		}
	}

	if allHealthy {
		h.Status = "healthy"
	} else if anyHealthy {
		h.Status = "degraded"
	} else {
		h.Status = "unhealthy"
	}
}
