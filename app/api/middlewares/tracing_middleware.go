package middlewares

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates an Echo middleware that instruments HTTP requests with OpenTelemetry tracing
// It should be placed after RequestIDMiddleware to have request IDs available
func TracingMiddleware() echo.MiddlewareFunc {
	// Use otelecho middleware which handles span creation and context propagation automatically
	return otelecho.Middleware(
		"anki-backend",
		otelecho.WithTracerProvider(otel.GetTracerProvider()),
		otelecho.WithPropagators(otel.GetTextMapPropagator()),
		otelecho.WithSkipper(func(c echo.Context) bool {
			// Skip tracing for health checks and metrics endpoints
			path := c.Request().URL.Path
			return path == "/health" || path == "/health/check" || path == "/metrics"
		}),
	)
}

// TracingMiddlewareWithCustomAttributes creates a tracing middleware with custom span attributes
// This adds request ID to spans for correlation with logs
func TracingMiddlewareWithCustomAttributes() echo.MiddlewareFunc {
	// Base middleware from otelecho handles HTTP instrumentation automatically
	baseMiddleware := otelecho.Middleware(
		"anki-backend",
		otelecho.WithTracerProvider(otel.GetTracerProvider()),
		otelecho.WithPropagators(otel.GetTextMapPropagator()),
		otelecho.WithSkipper(func(c echo.Context) bool {
			// Skip tracing for health checks and metrics endpoints
			path := c.Request().URL.Path
			return path == "/health" || path == "/health/check" || path == "/metrics"
		}),
	)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// Wrap the base middleware handler
		baseHandler := baseMiddleware(next)
		
		return func(c echo.Context) error {
			// Execute the base handler (which creates the span)
			err := baseHandler(c)

			// Add custom attributes to the span after execution
			span := trace.SpanFromContext(c.Request().Context())
			if span.IsRecording() {
				// Add request ID if available (for correlation with logs)
				requestID := GetRequestID(c.Request().Context())
				if requestID != "" {
					span.SetAttributes(
						attribute.String("request.id", requestID),
					)
				}

				// Record error if present (otelecho may not capture all errors)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
				}
			}

			return err
		}
	}
}

// GetTraceID extracts the trace ID from the context
func GetTraceID(ctx echo.Context) string {
	span := trace.SpanFromContext(ctx.Request().Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extracts the span ID from the context
func GetSpanID(ctx echo.Context) string {
	span := trace.SpanFromContext(ctx.Request().Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

