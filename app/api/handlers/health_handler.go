package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// HealthHandler handles health check requests
// This endpoint follows hexagonal architecture: Handler -> Service (IHealthService) -> Repositories (IDatabaseRepository, ICacheRepository)
// The service is implementation-agnostic and doesn't know about PostgreSQL or Redis specifics
type HealthHandler struct {
	healthService primary.IHealthService
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(healthService primary.IHealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// HealthCheck handles GET /health requests
func (h *HealthHandler) HealthCheck(c echo.Context) error {
	ctx := c.Request().Context()

	healthResp, err := h.healthService.CheckHealth(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check health",
		})
	}

	// Determine HTTP status code based on health status
	httpStatus := http.StatusOK
	if healthResp.Status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	} else if healthResp.Status == "degraded" {
		// For degraded (some components down but app still running), use 503 to indicate partial service availability
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, healthResp)
}
