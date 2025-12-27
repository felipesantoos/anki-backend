package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/felipesantos/anki-backend/config"
)

// TracingService manages OpenTelemetry tracing
type TracingService struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
	config         config.TracingConfig
}

// NewTracingService creates a new TracingService and initializes OpenTelemetry
func NewTracingService(cfg config.TracingConfig) (*TracingService, error) {
	if !cfg.Enabled {
		return &TracingService{config: cfg}, nil
	}

	service := &TracingService{
		config: cfg,
	}

	if err := service.setupTracerProvider(); err != nil {
		return nil, fmt.Errorf("failed to setup tracer provider: %w", err)
	}

	// Create tracer for this service
	service.tracer = service.tracerProvider.Tracer(
		cfg.ServiceName,
		trace.WithInstrumentationVersion("1.0.0"),
	)

	return service, nil
}

// setupTracerProvider configures and initializes the OpenTelemetry TracerProvider
func (ts *TracingService) setupTracerProvider() error {
	var exporters []sdktrace.SpanExporter

	// Setup Jaeger exporter if endpoint is configured
	if ts.config.JaegerEndpoint != "" {
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(ts.config.JaegerEndpoint),
		))
		if err != nil {
			return fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}
		exporters = append(exporters, jaegerExporter)
	}

	// Setup console exporter if enabled (useful for development)
	if ts.config.ConsoleEnabled {
		consoleExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return fmt.Errorf("failed to create console exporter: %w", err)
		}
		exporters = append(exporters, consoleExporter)
	}

	if len(exporters) == 0 {
		return fmt.Errorf("no exporters configured")
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", ts.config.ServiceName),
			attribute.String("service.version", "1.0.0"),
			attribute.String("deployment.environment", ts.config.Environment),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider options
	tpOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(ts.config.SampleRate)),
	}

	// Add batch processors for each exporter
	for _, exporter := range exporters {
		tpOptions = append(tpOptions, sdktrace.WithBatcher(exporter))
	}

	// Create tracer provider with all options
	tp := sdktrace.NewTracerProvider(tpOptions...)

	ts.tracerProvider = tp

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// GetTracer returns the tracer instance for creating spans
func (ts *TracingService) GetTracer() trace.Tracer {
	if ts.tracerProvider == nil {
		// Return a no-op tracer if tracing is disabled
		return trace.NewNoopTracerProvider().Tracer("noop")
	}
	return ts.tracer
}

// Shutdown gracefully shuts down the tracer provider
func (ts *TracingService) Shutdown(ctx context.Context) error {
	if ts.tracerProvider == nil {
		return nil
	}

	// Create context with timeout if not provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	return ts.tracerProvider.Shutdown(ctx)
}

// IsEnabled returns whether tracing is enabled
func (ts *TracingService) IsEnabled() bool {
	return ts.config.Enabled && ts.tracerProvider != nil
}

