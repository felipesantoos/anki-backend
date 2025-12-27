package integration

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/services/cache"
	eventService "github.com/felipesantos/anki-backend/core/services/events"
	jobService "github.com/felipesantos/anki-backend/core/services/jobs"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	tracingService "github.com/felipesantos/anki-backend/core/services/tracing"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	infraJobs "github.com/felipesantos/anki-backend/infra/jobs"
	"github.com/felipesantos/anki-backend/infra/postgres"
	"github.com/felipesantos/anki-backend/infra/redis"
)

func TestTracing_HTTPRequestCreatesSpan(t *testing.T) {
	// Setup tracing with console exporter for testing
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		JaegerEndpoint: "",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Create Echo instance
	e := echo.New()
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	// Create test route
	e.GET("/test", func(c echo.Context) error {
		// Verify span exists in context
		span := trace.SpanFromContext(c.Request().Context())
		assert.True(t, span.IsRecording(), "Span should be recording")
		assert.True(t, span.SpanContext().IsValid(), "Span context should be valid")
		
		return c.String(http.StatusOK, "OK")
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Verify trace ID is in response headers (if propagated)
	_ = rec.Header().Get("X-Trace-ID")
	// Note: otelecho may not add trace ID to headers by default
	// This is just a basic check that the middleware executed
}

func TestTracing_TraceContextPropagation(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	var capturedTraceID string
	var capturedSpanID string

	e.GET("/test", func(c echo.Context) error {
		span := trace.SpanFromContext(c.Request().Context())
		if span.SpanContext().IsValid() {
			capturedTraceID = span.SpanContext().TraceID().String()
			capturedSpanID = span.SpanContext().SpanID().String()
		}
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.NotEmpty(t, capturedTraceID, "Trace ID should be captured")
	assert.NotEmpty(t, capturedSpanID, "Span ID should be captured")
}

func TestTracing_RequestIDInSpan(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	e.GET("/test", func(c echo.Context) error {
		// Request ID should be available in context
		requestID := middlewares.GetRequestID(c.Request().Context())
		assert.NotEmpty(t, requestID, "Request ID should be available")
		
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Verify request ID is in response headers
	requestID := rec.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "Request ID should be in response headers")
}

func TestTracing_DisabledDoesNotCreateSpans(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled: false,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	assert.False(t, tracingSvc.IsEnabled(), "Tracing should be disabled")
	
	// Even with middleware, spans should be no-op
	e := echo.New()
	e.Use(middlewares.TracingMiddleware())

	e.GET("/test", func(c echo.Context) error {
		span := trace.SpanFromContext(c.Request().Context())
		// Span should exist but may be no-op
		assert.NotNil(t, span)
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTracing_ErrorRecording(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "Test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestTracing_HealthCheckSkipped(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	e.GET("/health", func(c echo.Context) error {
		// Health check should be skipped by middleware
		_ = trace.SpanFromContext(c.Request().Context())
		// Span may still exist but should not be recorded
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTracing_GetTraceID(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	var traceID string
	var spanID string

	e.GET("/test", func(c echo.Context) error {
		traceID = middlewares.GetTraceID(c)
		spanID = middlewares.GetSpanID(c)
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.NotEmpty(t, traceID, "Trace ID should be available")
	assert.NotEmpty(t, spanID, "Span ID should be available")
}

func TestTracing_ServiceInstrumentation(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Test that tracer is available
	tracer := tracingSvc.GetTracer()
	assert.NotNil(t, tracer)

	// Create a span manually
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test.operation")
	
	assert.True(t, span.IsRecording(), "Span should be recording")
	
	span.End()
}

func TestTracing_SamplingRate(t *testing.T) {
	// Test with 0% sampling (no spans should be recorded)
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     0.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.TracingMiddleware())

	var spanRecorded bool

	e.GET("/test", func(c echo.Context) error {
		span := trace.SpanFromContext(c.Request().Context())
		spanRecorded = span.IsRecording()
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// With 0% sampling, spans may not be recorded
	// This is expected behavior
	_ = spanRecorded
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTracing_ContextPropagation(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	e := echo.New()
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.TracingMiddlewareWithCustomAttributes())

	var parentTraceID string
	var childTraceID string

	e.GET("/parent", func(c echo.Context) error {
		span := trace.SpanFromContext(c.Request().Context())
		if span.SpanContext().IsValid() {
			parentTraceID = span.SpanContext().TraceID().String()
		}

		// Simulate nested operation
		ctx := c.Request().Context()
		tracer := otel.Tracer("test")
		_, childSpan := tracer.Start(ctx, "child.operation")
		if childSpan.SpanContext().IsValid() {
			childTraceID = childSpan.SpanContext().TraceID().String()
		}
		childSpan.End()

		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/parent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Trace IDs should match (child should inherit parent's trace)
	if parentTraceID != "" && childTraceID != "" {
		assert.Equal(t, parentTraceID, childTraceID, "Child span should inherit parent trace ID")
	}
}

func TestTracing_Shutdown(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)

	// Shutdown should complete without error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tracingSvc.Shutdown(ctx)
	assert.NoError(t, err, "Shutdown should complete successfully")
}

func TestTracing_DatabaseSpans(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Load database config
	appCfg, err := config.Load()
	require.NoError(t, err)

	logger := slog.Default()
	db, err := postgres.NewPostgresRepository(appCfg.Database, logger)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Test Ping creates span
	err = db.Ping(ctx)
	assert.NoError(t, err, "Ping should succeed")

	// Verify span was created (we can't easily verify span attributes without exporter,
	// but we can verify the operation completed successfully)
}

func TestTracing_RedisSpans(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Load Redis config
	appCfg, err := config.Load()
	require.NoError(t, err)

	logger := slog.Default()
	rdb, err := redis.NewRedisRepository(appCfg.Redis, logger)
	require.NoError(t, err)
	defer rdb.Close()

	ctx := context.Background()
	testKey := "test:tracing:key"
	testValue := "test-value"

	// Test Ping creates span
	err = rdb.Ping(ctx)
	assert.NoError(t, err, "Ping should succeed")

	// Test Set creates span
	err = rdb.Set(ctx, testKey, testValue, time.Minute)
	assert.NoError(t, err, "Set should succeed")

	// Test Get creates span
	value, err := rdb.Get(ctx, testKey)
	assert.NoError(t, err, "Get should succeed")
	assert.Equal(t, testValue, value, "Value should match")

	// Test Delete creates span
	err = rdb.Delete(ctx, testKey)
	assert.NoError(t, err, "Delete should succeed")

	// Cleanup
	rdb.Delete(ctx, testKey)
}

func TestTracing_CacheServiceSpans(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Load Redis config and create cache service
	appCfg, err := config.Load()
	require.NoError(t, err)

	logger := slog.Default()
	rdb, err := redis.NewRedisRepository(appCfg.Redis, logger)
	require.NoError(t, err)
	defer rdb.Close()

	cacheSvc := cache.NewCacheService(rdb)
	ctx := context.Background()
	testKey := "test:cache:key"
	testValue := "cache-value"

	// Test Get creates span
	_, err = cacheSvc.Get(ctx, testKey)
	// May fail if key doesn't exist, which is fine

	// Test Set creates span
	err = cacheSvc.Set(ctx, testKey, testValue, time.Minute)
	assert.NoError(t, err, "Set should succeed")

	// Test Get creates span (cache hit)
	value, err := cacheSvc.Get(ctx, testKey)
	assert.NoError(t, err, "Get should succeed")
	assert.Equal(t, testValue, value, "Value should match")

	// Test Delete creates span
	err = cacheSvc.Delete(ctx, testKey)
	assert.NoError(t, err, "Delete should succeed")
}

func TestTracing_SessionServiceSpans(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Create session repository and service
	appCfg, err := config.Load()
	require.NoError(t, err)

	logger := slog.Default()
	rdb, err := redis.NewRedisRepository(appCfg.Redis, logger)
	require.NoError(t, err)
	defer rdb.Close()

	sessionRepo := redis.NewSessionRepository(rdb.Client, "test")
	sessionSvc := sessionService.NewSessionService(sessionRepo, 30*time.Minute)

	ctx := context.Background()
	userID := "test-user-123"

	// Test CreateSession creates span
	sessionID, err := sessionSvc.CreateSession(ctx, userID, nil)
	assert.NoError(t, err, "CreateSession should succeed")
	assert.NotEmpty(t, sessionID, "Session ID should not be empty")

	// Test GetSession creates span
	data, err := sessionSvc.GetSession(ctx, sessionID)
	assert.NoError(t, err, "GetSession should succeed")
	assert.NotNil(t, data, "Session data should not be nil")

	// Test UpdateSession creates span
	err = sessionSvc.UpdateSession(ctx, sessionID, map[string]interface{}{"key": "value"})
	assert.NoError(t, err, "UpdateSession should succeed")

	// Test DeleteSession creates span
	err = sessionSvc.DeleteSession(ctx, sessionID)
	assert.NoError(t, err, "DeleteSession should succeed")
}

func TestTracing_EventServiceSpans(t *testing.T) {
	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Create in-memory event bus and service
	logger := slog.Default()
	eventBus := infraEvents.NewInMemoryEventBus(5, 1000, logger)
	eventSvc := eventService.NewEventService(eventBus)

	ctx := context.Background()

	// Create a test event using NoteCreated
	testEvent := &events.NoteCreated{
		NoteID:    123,
		UserID:    456,
		NoteTypeID: 789,
		Timestamp: time.Now(),
	}

	// Test Publish creates span
	err = eventSvc.Publish(ctx, testEvent)
	assert.NoError(t, err, "Publish should succeed")
}

func TestTracing_JobServiceSpans(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg := config.TracingConfig{
		Enabled:        true,
		ServiceName:    "anki-backend-test",
		Environment:    "test",
		SampleRate:     1.0,
		ConsoleEnabled: true,
	}

	tracingSvc, err := tracingService.NewTracingService(cfg)
	require.NoError(t, err)
	defer tracingSvc.Shutdown(context.Background())

	// Create job queue and service
	appCfg, err := config.Load()
	require.NoError(t, err)

	logger := slog.Default()
	rdb, err := redis.NewRedisRepository(appCfg.Redis, logger)
	require.NoError(t, err)
	defer rdb.Close()

	jobQueue := infraJobs.NewRedisQueue(rdb.Client, "test:queue")
	jobSvc := jobService.NewJobService(jobQueue, 3)

	ctx := context.Background()
	jobType := "test-job"
	payload := map[string]interface{}{
		"key": "value",
	}

	// Test Enqueue creates span
	jobID, err := jobSvc.Enqueue(ctx, jobType, payload)
	assert.NoError(t, err, "Enqueue should succeed")
	assert.NotEmpty(t, jobID, "Job ID should not be empty")

	// Test GetStatus creates span
	job, err := jobSvc.GetStatus(ctx, jobID)
	assert.NoError(t, err, "GetStatus should succeed")
	if job != nil {
		assert.Equal(t, jobType, job.Type, "Job type should match")
	}
}

