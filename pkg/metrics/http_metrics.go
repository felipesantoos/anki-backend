package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// HTTPMetrics holds all HTTP-related Prometheus metrics
type HTTPMetrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	RequestSize       *prometheus.HistogramVec
	ResponseSize      *prometheus.HistogramVec
}

// NewHTTPMetrics creates a new HTTPMetrics instance with all HTTP metrics configured
func NewHTTPMetrics() *HTTPMetrics {
	return &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status_code"},
		),
		RequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to 1GB
			},
			[]string{"method", "path"},
		),
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to 1GB
			},
			[]string{"method", "path", "status_code"},
		),
	}
}

// Register registers all HTTP metrics with the given Prometheus registry
func (h *HTTPMetrics) Register(registry *prometheus.Registry) error {
	registry.MustRegister(h.RequestsTotal)
	registry.MustRegister(h.RequestDuration)
	registry.MustRegister(h.RequestSize)
	registry.MustRegister(h.ResponseSize)
	return nil
}

// RecordRequest records a complete HTTP request with all metrics
func (h *HTTPMetrics) RecordRequest(method, path, statusCode string, duration float64, requestSize, responseSize int64) {
	h.RequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	h.RequestDuration.WithLabelValues(method, path, statusCode).Observe(duration)
	if requestSize > 0 {
		h.RequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	}
	if responseSize > 0 {
		h.ResponseSize.WithLabelValues(method, path, statusCode).Observe(float64(responseSize))
	}
}

// RegisterHTTPMetrics registers HTTP metrics with the given Prometheus registry
// This function is kept for backward compatibility but is deprecated.
// Use NewHTTPMetrics() and Register() instead.
func RegisterHTTPMetrics(registry *prometheus.Registry) error {
	httpMetrics := NewHTTPMetrics()
	return httpMetrics.Register(registry)
}
