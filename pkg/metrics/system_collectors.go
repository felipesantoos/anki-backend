package metrics

import (
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// DatabaseCollector collects metrics from a PostgreSQL database connection pool
type DatabaseCollector struct {
	db                *sql.DB
	systemMetrics     *SystemMetrics
}

// NewDatabaseCollector creates a new DatabaseCollector
func NewDatabaseCollector(db *sql.DB, systemMetrics *SystemMetrics) *DatabaseCollector {
	return &DatabaseCollector{
		db:            db,
		systemMetrics: systemMetrics,
	}
}

// Describe implements prometheus.Collector
func (dc *DatabaseCollector) Describe(ch chan<- *prometheus.Desc) {
	if dc.systemMetrics != nil {
		dc.systemMetrics.DatabaseConnectionsActive.Describe(ch)
	}
}

// Collect implements prometheus.Collector
func (dc *DatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	if dc.db == nil || dc.systemMetrics == nil {
		return
	}

	stats := dc.db.Stats()
	dc.systemMetrics.DatabaseConnectionsActive.Set(float64(stats.OpenConnections))
	dc.systemMetrics.DatabaseConnectionsActive.Collect(ch)
}

// RedisCollector collects metrics from a Redis client
type RedisCollector struct {
	client        *redis.Client
	systemMetrics *SystemMetrics
}

// NewRedisCollector creates a new RedisCollector
func NewRedisCollector(client *redis.Client, systemMetrics *SystemMetrics) *RedisCollector {
	return &RedisCollector{
		client:        client,
		systemMetrics: systemMetrics,
	}
}

// Describe implements prometheus.Collector
func (rc *RedisCollector) Describe(ch chan<- *prometheus.Desc) {
	if rc.systemMetrics != nil {
		rc.systemMetrics.RedisConnectionsActive.Describe(ch)
		rc.systemMetrics.RedisPoolSize.Describe(ch)
	}
}

// Collect implements prometheus.Collector
func (rc *RedisCollector) Collect(ch chan<- prometheus.Metric) {
	if rc.client == nil || rc.systemMetrics == nil {
		return
	}

	poolStats := rc.client.PoolStats()
	rc.systemMetrics.RedisConnectionsActive.Set(float64(poolStats.TotalConns))
	rc.systemMetrics.RedisPoolSize.Set(float64(poolStats.TotalConns)) // Total connections in pool
	
	rc.systemMetrics.RedisConnectionsActive.Collect(ch)
	rc.systemMetrics.RedisPoolSize.Collect(ch)
}

// RegisterDatabaseCollector registers a DatabaseCollector with the given registry
func RegisterDatabaseCollector(registry *prometheus.Registry, db *sql.DB, systemMetrics *SystemMetrics) error {
	collector := NewDatabaseCollector(db, systemMetrics)
	return registry.Register(collector)
}

// RegisterRedisCollector registers a RedisCollector with the given registry
func RegisterRedisCollector(registry *prometheus.Registry, client *redis.Client, systemMetrics *SystemMetrics) error {
	collector := NewRedisCollector(client, systemMetrics)
	return registry.Register(collector)
}

// ObserveDatabaseQueryDuration records the duration of a database query
func ObserveDatabaseQueryDuration(systemMetrics *SystemMetrics, operation string, duration time.Duration) {
	if systemMetrics != nil && systemMetrics.DatabaseQueryDuration != nil {
		systemMetrics.DatabaseQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
	}
}

// ObserveRedisCommandDuration records the duration of a Redis command
func ObserveRedisCommandDuration(systemMetrics *SystemMetrics, command string, duration time.Duration) {
	if systemMetrics != nil && systemMetrics.RedisCommandDuration != nil {
		systemMetrics.RedisCommandDuration.WithLabelValues(command).Observe(duration.Seconds())
	}
}



