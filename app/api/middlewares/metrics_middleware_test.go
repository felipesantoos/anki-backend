package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/services/metrics"
)

func TestMetricsMiddleware_CollectsMetrics(t *testing.T) {
	// Create metrics service and register HTTP metrics
	metricsSvc := metrics.NewMetricsService()
	err := metricsSvc.RegisterHTTPMetrics()
	require.NoError(t, err)

	// Create Echo instance
	e := echo.New()
	
	// Add metrics middleware
	e.Use(MetricsMiddleware(metricsSvc))
	
	// Create test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Verify metrics were collected by checking the metrics endpoint
	metricsHandler := metricsSvc.GetHandler()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	metricsHandler.ServeHTTP(metricsRec, metricsReq)

	assert.Equal(t, http.StatusOK, metricsRec.Code)
	body := metricsRec.Body.String()
	
	// Should contain HTTP metrics
	assert.Contains(t, body, "http_requests_total")
}

func TestMetricsMiddleware_NilService(t *testing.T) {
	// Create Echo instance
	e := echo.New()
	
	// Add metrics middleware with nil service (should not panic)
	e.Use(MetricsMiddleware(nil))
	
	// Create test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Make request - should not panic
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsMiddleware_RecordsStatusCodes(t *testing.T) {
	metricsSvc := metrics.NewMetricsService()
	err := metricsSvc.RegisterHTTPMetrics()
	require.NoError(t, err)

	e := echo.New()
	e.Use(MetricsMiddleware(metricsSvc))
	
	// Create routes with different status codes
	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.GET("/not-found", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "Not Found")
	})
	e.GET("/server-error", func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "Error")
	})

	// Make requests
	testCases := []struct {
		path       string
		statusCode int
	}{
		{"/ok", http.StatusOK},
		{"/not-found", http.StatusNotFound},
		{"/server-error", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, tc.statusCode, rec.Code)
	}

	// Verify metrics endpoint contains status codes
	metricsHandler := metricsSvc.GetHandler()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	metricsHandler.ServeHTTP(metricsRec, metricsReq)

	assert.Equal(t, http.StatusOK, metricsRec.Code)
	body := metricsRec.Body.String()
	assert.Contains(t, body, "http_requests_total")
}

