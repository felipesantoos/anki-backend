package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// SystemMetrics holds all system-related Prometheus metrics
type SystemMetrics struct {
	DatabaseConnectionsActive prometheus.Gauge
	DatabaseQueryDuration     *prometheus.HistogramVec
	RedisConnectionsActive    prometheus.Gauge
	RedisCommandDuration      *prometheus.HistogramVec
	RedisPoolSize             prometheus.Gauge
}

// NewSystemMetrics creates a new SystemMetrics instance with all system metrics configured
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		DatabaseConnectionsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Number of active database connections",
			},
		),
		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"operation"}, // e.g., "select", "insert", "update", "delete"
		),
		RedisConnectionsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connections_active",
				Help: "Number of active Redis connections",
			},
		),
		RedisCommandDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_command_duration_seconds",
				Help:    "Redis command duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"command"}, // e.g., "get", "set", "hget", "hset"
		),
		RedisPoolSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_pool_size",
				Help: "Size of the Redis connection pool",
			},
		),
	}
}

// Register registers all system metrics with the given Prometheus registry
func (s *SystemMetrics) Register(registry *prometheus.Registry) error {
	registry.MustRegister(s.DatabaseConnectionsActive)
	registry.MustRegister(s.DatabaseQueryDuration)
	registry.MustRegister(s.RedisConnectionsActive)
	registry.MustRegister(s.RedisCommandDuration)
	registry.MustRegister(s.RedisPoolSize)
	return nil
}

// RegisterSystemMetrics registers system metrics with the given Prometheus registry
// This function is kept for backward compatibility but is deprecated.
// Use NewSystemMetrics() and Register() instead.
func RegisterSystemMetrics(registry *prometheus.Registry) error {
	systemMetrics := NewSystemMetrics()
	return systemMetrics.Register(registry)
}
