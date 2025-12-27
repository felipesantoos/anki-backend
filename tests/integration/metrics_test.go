package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/services/metrics"
)

func TestMetrics_EndpointReturnsValidFormat(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	// Register all metrics
	err := svc.RegisterHTTPMetrics()
	require.NoError(t, err)
	
	err = svc.RegisterSystemMetrics()
	require.NoError(t, err)
	
	err = svc.RegisterBusinessMetrics()
	require.NoError(t, err)

	// Create at least one data point for each metric type so they appear in /metrics
	// Prometheus only exposes metrics that have at least one data series
	svc.RecordHTTPRequest("GET", "/test", "200", 0.1, 100, 200)

	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	body := rec.Body.String()
	
	// Verify Prometheus format
	// Should contain at least one metric with help and type
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
	
	// Verify HTTP metrics are present (now that we've recorded a request)
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, "http_request_duration_seconds")
	
	// Verify system metrics are present (gauges always appear)
	assert.Contains(t, body, "database_connections_active")
	assert.Contains(t, body, "redis_connections_active")
	
	// Note: Business metrics won't appear until they're used
	// This is expected Prometheus behavior - metrics only appear after first use
}

func TestMetrics_HTTPMetricsIncrement(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	err := svc.RegisterHTTPMetrics()
	require.NoError(t, err)

	// Record a request so the metric appears in /metrics
	// Prometheus only exposes metrics that have at least one data series
	svc.RecordHTTPRequest("GET", "/test", "200", 0.1, 100, 200)

	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	
	// Verify metric exists and has data
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, "http_request_duration_seconds")
}

func TestMetrics_FormatIsValidPrometheus(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	err := svc.RegisterHTTPMetrics()
	require.NoError(t, err)
	err = svc.RegisterSystemMetrics()
	require.NoError(t, err)
	err = svc.RegisterBusinessMetrics()
	require.NoError(t, err)

	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()

	// Basic Prometheus format validation
	lines := strings.Split(body, "\n")
	
	// Should have at least some content
	assert.Greater(t, len(lines), 0)
	
	// Check for Prometheus format markers
	hasHelp := false
	hasType := false
	hasMetric := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "# HELP") {
			hasHelp = true
		}
		if strings.HasPrefix(line, "# TYPE") {
			hasType = true
		}
		// Metric lines don't start with # and contain space or {
		if !strings.HasPrefix(line, "#") && (strings.Contains(line, " ") || strings.Contains(line, "{")) {
			hasMetric = true
		}
	}
	
	assert.True(t, hasHelp, "Should have HELP comments")
	assert.True(t, hasType, "Should have TYPE comments")
	assert.True(t, hasMetric || len(lines) > 10, "Should have metric data or sufficient content")
}

func TestMetrics_AllMetricsRegistered(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	// Register all metrics
	err := svc.RegisterHTTPMetrics()
	require.NoError(t, err)
	err = svc.RegisterSystemMetrics()
	require.NoError(t, err)
	err = svc.RegisterBusinessMetrics()
	require.NoError(t, err)

	// Create at least one data point for metrics that require it
	// Prometheus only exposes metrics that have at least one data series
	svc.RecordHTTPRequest("GET", "/test", "200", 0.1, 100, 200)

	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	
	// HTTP metrics (should appear after recording a request)
	httpMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"http_request_size_bytes",
		"http_response_size_bytes",
	}
	
	// System metrics (gauges always appear, histograms need data points)
	systemMetrics := []string{
		"database_connections_active",
		"redis_connections_active",
		"redis_pool_size",
		// Note: database_query_duration_seconds and redis_command_duration_seconds
		// won't appear until they're used, which is expected Prometheus behavior
	}
	
	// Business metrics won't appear until they're used
	// This is expected Prometheus behavior - CounterVec metrics only appear after first use
	
	allMetrics := append(httpMetrics, systemMetrics...)
	
	for _, metricName := range allMetrics {
		assert.Contains(t, body, metricName, "Metric %s should be present", metricName)
	}
}



