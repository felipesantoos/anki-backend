package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterMetricsRoutes registers the metrics endpoint on the Router
func (r *Router) RegisterMetricsRoutes() {
	metricsService := dicontainer.GetMetricsService()
	if metricsService == nil {
		return
	}

	cfg := dicontainer.GetConfig()
	path := cfg.Metrics.Path
	if path == "" {
		path = "/metrics"
	}

	r.echo.GET(path, echo.WrapHandler(metricsService.GetHandler()))
}



