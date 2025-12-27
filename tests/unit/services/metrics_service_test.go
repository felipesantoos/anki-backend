package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/services/metrics"
)

func TestMetricsService_NewMetricsService(t *testing.T) {
	svc := metrics.NewMetricsService()
	require.NotNil(t, svc)

	// Verify handler is created
	handler := svc.GetHandler()
	require.NotNil(t, handler)
}

func TestMetricsService_RegisterHTTPMetrics(t *testing.T) {
	svc := metrics.NewMetricsService()
	err := svc.RegisterHTTPMetrics()
	require.NoError(t, err)

	// Verify metrics are registered by checking handler returns data
	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "http_requests_total")
}

func TestMetricsService_RegisterSystemMetrics(t *testing.T) {
	svc := metrics.NewMetricsService()
	err := svc.RegisterSystemMetrics()
	require.NoError(t, err)

	// Verify metrics are registered
	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "database_connections_active")
}

func TestMetricsService_RegisterBusinessMetrics(t *testing.T) {
	svc := metrics.NewMetricsService()
	err := svc.RegisterBusinessMetrics()
	require.NoError(t, err)

	// Verify metrics are registered
	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "decks_created_total")
}

func TestMetricsService_GetHandler(t *testing.T) {
	svc := metrics.NewMetricsService()
	handler := svc.GetHandler()
	require.NotNil(t, handler)

	// Test handler returns valid Prometheus format
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestMetricsService_IncrementCounter(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	// Register a test counter manually
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter_total",
			Help: "Test counter",
		},
		[]string{"label1"},
	)
	svc.GetRegistry().MustRegister(counter)

	// This should not panic even if metric is not registered via IncrementCounter
	// (since we're testing the method with a manually registered counter)
	svc.IncrementCounter("test_counter_total", map[string]string{"label1": "value1"})

	// Verify by scraping metrics
	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Note: The counter might be 0 if not actually incremented through the service
	// This test mainly verifies the method doesn't panic
}

func TestMetricsService_ObserveHistogram(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	// Register a test histogram manually
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "test_histogram",
			Help: "Test histogram",
		},
		[]string{"label1"},
	)
	svc.GetRegistry().MustRegister(histogram)

	// This should not panic even if metric is not registered via ObserveHistogram
	svc.ObserveHistogram("test_histogram", 1.5, map[string]string{"label1": "value1"})

	// Verify by scraping metrics
	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsService_SetGauge(t *testing.T) {
	svc := metrics.NewMetricsService()
	
	// Register a test gauge manually
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_gauge",
			Help: "Test gauge",
		},
		[]string{"label1"},
	)
	svc.GetRegistry().MustRegister(gauge)

	// This should not panic even if metric is not registered via SetGauge
	svc.SetGauge("test_gauge", 42.0, map[string]string{"label1": "value1"})

	// Verify by scraping metrics
	handler := svc.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}



