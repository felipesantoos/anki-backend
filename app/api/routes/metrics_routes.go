package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// RegisterMetricsRoutes registers the metrics endpoint
func RegisterMetricsRoutes(e *echo.Echo, metricsService primary.IMetricsService, path string) {
	if metricsService == nil {
		return
	}

	e.GET(path, echo.WrapHandler(metricsService.GetHandler()))
}



