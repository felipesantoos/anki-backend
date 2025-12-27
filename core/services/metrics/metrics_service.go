package metrics

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	metricsPackage "github.com/felipesantos/anki-backend/pkg/metrics"
)

// MetricsService implements IMetricsService
// It manages a custom Prometheus registry and provides methods to record metrics
type MetricsService struct {
	registry *prometheus.Registry
	mu       sync.RWMutex
	
	// Store metrics as structs instead of global variables
	httpMetrics     *metricsPackage.HTTPMetrics
	systemMetrics   *metricsPackage.SystemMetrics
	businessMetrics *metricsPackage.BusinessMetrics
	
	// Keep maps for generic methods (IncrementCounter, etc.)
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	gauges     map[string]*prometheus.GaugeVec
	
	// Track which collectors have been registered
	databaseCollectorRegistered bool
	redisCollectorRegistered     bool
}

// NewMetricsService creates a new MetricsService with a custom Prometheus registry
func NewMetricsService() *MetricsService {
	return &MetricsService{
		registry:   prometheus.NewRegistry(),
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
	}
}

// RegisterHTTPMetrics registers HTTP-related metrics
func (m *MetricsService) RegisterHTTPMetrics() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.httpMetrics != nil {
		// Already registered
		return nil
	}
	
	m.httpMetrics = metricsPackage.NewHTTPMetrics()
	if err := m.httpMetrics.Register(m.registry); err != nil {
		return fmt.Errorf("failed to register HTTP metrics: %w", err)
	}
	return nil
}

// RegisterSystemMetrics registers system metrics
func (m *MetricsService) RegisterSystemMetrics() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.systemMetrics != nil {
		// Already registered
		return nil
	}
	
	m.systemMetrics = metricsPackage.NewSystemMetrics()
	if err := m.systemMetrics.Register(m.registry); err != nil {
		return fmt.Errorf("failed to register system metrics: %w", err)
	}
	return nil
}

// RegisterBusinessMetrics registers business domain metrics
func (m *MetricsService) RegisterBusinessMetrics() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.businessMetrics != nil {
		// Already registered
		return nil
	}
	
	m.businessMetrics = metricsPackage.NewBusinessMetrics()
	if err := m.businessMetrics.Register(m.registry); err != nil {
		return fmt.Errorf("failed to register business metrics: %w", err)
	}
	return nil
}

// RecordHTTPRequest records a complete HTTP request with all relevant metrics
func (m *MetricsService) RecordHTTPRequest(method, path, statusCode string, duration float64, requestSize, responseSize int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.httpMetrics == nil {
		// HTTP metrics not registered, ignore silently
		return
	}
	
	m.httpMetrics.RecordRequest(method, path, statusCode, duration, requestSize, responseSize)
}

// IncrementCounter increments a counter metric by 1
func (m *MetricsService) IncrementCounter(name string, labels map[string]string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counter, exists := m.counters[name]
	if !exists {
		// Counter not registered, ignore silently or log warning
		// In a production system, you might want to log this
		return
	}

	counter.With(prometheus.Labels(labels)).Inc()
}

// ObserveHistogram records a value in a histogram metric
func (m *MetricsService) ObserveHistogram(name string, value float64, labels map[string]string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	histogram, exists := m.histograms[name]
	if !exists {
		// Histogram not registered, ignore silently
		return
	}

	histogram.With(prometheus.Labels(labels)).Observe(value)
}

// SetGauge sets a gauge metric to a specific value
func (m *MetricsService) SetGauge(name string, value float64, labels map[string]string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gauge, exists := m.gauges[name]
	if !exists {
		// Gauge not registered, ignore silently
		return
	}

	gauge.With(prometheus.Labels(labels)).Set(value)
}

// GetHandler returns the HTTP handler for the /metrics endpoint
func (m *MetricsService) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// getOrCreateCounter gets an existing counter or creates a new one
// This is a helper method for internal use
func (m *MetricsService) getOrCreateCounter(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if counter, exists := m.counters[opts.Name]; exists {
		return counter
	}

	counter := prometheus.NewCounterVec(opts, labelNames)
	m.registry.MustRegister(counter)
	m.counters[opts.Name] = counter
	return counter
}

// getOrCreateHistogram gets an existing histogram or creates a new one
// This is a helper method for internal use
func (m *MetricsService) getOrCreateHistogram(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if histogram, exists := m.histograms[opts.Name]; exists {
		return histogram
	}

	histogram := prometheus.NewHistogramVec(opts, labelNames)
	m.registry.MustRegister(histogram)
	m.histograms[opts.Name] = histogram
	return histogram
}

// getOrCreateGauge gets an existing gauge or creates a new one
// This is a helper method for internal use
func (m *MetricsService) getOrCreateGauge(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	m.mu.Lock()
	defer m.mu.Unlock()

	if gauge, exists := m.gauges[opts.Name]; exists {
		return gauge
	}

	gauge := prometheus.NewGaugeVec(opts, labelNames)
	m.registry.MustRegister(gauge)
	m.gauges[opts.Name] = gauge
	return gauge
}

// GetRegistry returns the underlying Prometheus registry
// This is useful for registering custom metrics directly
func (m *MetricsService) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RegisterDatabaseCollector registers a database collector for system metrics
func (m *MetricsService) RegisterDatabaseCollector(db *sql.DB) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.databaseCollectorRegistered {
		// Already registered, skip
		return nil
	}
	
	if m.systemMetrics == nil {
		return fmt.Errorf("system metrics must be registered before registering database collector")
	}
	
	if err := metricsPackage.RegisterDatabaseCollector(m.registry, db, m.systemMetrics); err != nil {
		return err
	}
	
	m.databaseCollectorRegistered = true
	return nil
}

// RegisterRedisCollector registers a Redis collector for system metrics
func (m *MetricsService) RegisterRedisCollector(client *redis.Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.redisCollectorRegistered {
		// Already registered, skip
		return nil
	}
	
	if m.systemMetrics == nil {
		return fmt.Errorf("system metrics must be registered before registering redis collector")
	}
	
	if err := metricsPackage.RegisterRedisCollector(m.registry, client, m.systemMetrics); err != nil {
		return err
	}
	
	m.redisCollectorRegistered = true
	return nil
}

// Ensure MetricsService implements primary.IMetricsService
var _ primary.IMetricsService = (*MetricsService)(nil)

