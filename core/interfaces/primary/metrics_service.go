package primary

import "net/http"

// IMetricsService defines the interface for metrics collection and exposure
// Following hexagonal architecture, this interface is in the primary (driving) port
type IMetricsService interface {
	// RegisterHTTPMetrics registers HTTP-related metrics (requests, duration, etc.)
	RegisterHTTPMetrics() error

	// RegisterSystemMetrics registers system metrics (database, Redis, etc.)
	RegisterSystemMetrics() error

	// RegisterBusinessMetrics registers business domain metrics
	RegisterBusinessMetrics() error

	// IncrementCounter increments a counter metric by 1
	// name: metric name (e.g., "http_requests_total")
	// labels: key-value pairs for labels (e.g., map[string]string{"method": "GET", "status": "200"})
	IncrementCounter(name string, labels map[string]string)

	// ObserveHistogram records a value in a histogram metric
	// name: metric name (e.g., "http_request_duration_seconds")
	// value: the value to observe (e.g., 0.123)
	// labels: key-value pairs for labels
	ObserveHistogram(name string, value float64, labels map[string]string)

	// SetGauge sets a gauge metric to a specific value
	// name: metric name (e.g., "database_connections_active")
	// value: the gauge value
	// labels: key-value pairs for labels
	SetGauge(name string, value float64, labels map[string]string)

	// RecordHTTPRequest records a complete HTTP request with all relevant metrics
	// method: HTTP method (e.g., "GET", "POST")
	// path: normalized HTTP path (e.g., "/api/users/:id")
	// statusCode: HTTP status code as string (e.g., "200", "404")
	// duration: request duration in seconds
	// requestSize: request size in bytes
	// responseSize: response size in bytes
	RecordHTTPRequest(method, path, statusCode string, duration float64, requestSize, responseSize int64)

	// GetHandler returns the HTTP handler for the /metrics endpoint
	// This handler exposes all registered metrics in Prometheus format
	GetHandler() http.Handler
}


