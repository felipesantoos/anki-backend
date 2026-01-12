package routes

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/dicontainer"
	"github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// Router handles the registration of all API routes and middlewares
type Router struct {
	echo   *echo.Echo
	cfg    *config.Config
	jwtSvc *jwt.JWTService
	rdb    *redis.RedisRepository
}

// NewRouter creates a new Router instance
func NewRouter(e *echo.Echo, cfg *config.Config, jwtSvc *jwt.JWTService, rdb *redis.RedisRepository) *Router {
	// Set up validator early so it's available even if routes are registered individually
	e.Validator = middlewares.NewCustomValidator()
	
	return &Router{
		echo:   e,
		cfg:    cfg,
		jwtSvc: jwtSvc,
		rdb:    rdb,
	}
}

// Init initializes the router with middlewares and all routes
func (r *Router) Init() {
	r.setupMiddlewares()
	r.RegisterAll()
}

// setupMiddlewares configures global middlewares
func (r *Router) setupMiddlewares() {
	r.echo.HideBanner = true
	r.echo.HTTPErrorHandler = middlewares.CustomHTTPErrorHandler

	// Validator is already set up in NewRouter, but ensure it's set here too
	if r.echo.Validator == nil {
		r.echo.Validator = middlewares.NewCustomValidator()
	}

	r.echo.Use(echoMiddleware.Recover())
	r.echo.Use(middlewares.CORSMiddleware(r.cfg.CORS))
	r.echo.Use(middlewares.RequestIDMiddleware())
	
	if r.cfg.Tracing.Enabled {
		r.echo.Use(middlewares.TracingMiddlewareWithCustomAttributes())
	}
	
	if r.cfg.Metrics.Enabled && r.cfg.Metrics.EnableHTTPMetrics {
		metricsSvc := dicontainer.GetMetricsService()
		r.echo.Use(middlewares.MetricsMiddleware(metricsSvc))
	}
	
	r.echo.Use(middlewares.RateLimitingMiddleware(r.cfg.RateLimit, r.rdb.Client))
}

// RegisterAll registers all routes in the application
func (r *Router) RegisterAll() {
	r.RegisterSwaggerRoutes()
	r.RegisterHealthRoutes()
	r.RegisterMetricsRoutes()
	r.RegisterAuthRoutes()
	r.RegisterStudyRoutes()
	r.RegisterContentRoutes()
	r.RegisterUserRoutes()
	r.RegisterSystemRoutes()
	r.RegisterCommunityRoutes()
	r.RegisterSearchRoutes()
	r.RegisterMaintenanceRoutes()
}

// RegisterSwaggerRoutes registers the Swagger documentation routes
func (r *Router) RegisterSwaggerRoutes() {
	r.echo.GET("/swagger/*", echoSwagger.WrapHandler)
}

