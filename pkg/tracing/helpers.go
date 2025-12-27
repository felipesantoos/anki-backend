package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan creates a new span as a child of the span in the context
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("anki-backend")
	return tracer.Start(ctx, name, opts...)
}

// AddSpanAttributes adds attributes to a span
func AddSpanAttributes(span trace.Span, attrs map[string]string) {
	if !span.IsRecording() {
		return
	}

	attributes := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		attributes = append(attributes, attribute.String(k, v))
	}
	span.SetAttributes(attributes...)
}

// AddSpanAttributesKV adds key-value attributes to a span
func AddSpanAttributesKV(span trace.Span, attrs ...attribute.KeyValue) {
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(attrs...)
}

// RecordError records an error in a span
func RecordError(span trace.Span, err error) {
	if err == nil || !span.IsRecording() {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// GetTraceID extracts the trace ID from the context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extracts the span ID from the context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// IsTracingEnabled checks if tracing is enabled in the context
func IsTracingEnabled(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	return span.IsRecording()
}

