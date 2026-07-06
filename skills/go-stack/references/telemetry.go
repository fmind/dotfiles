package <slug>

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
)

// SetupOTel installs a global OpenTelemetry tracer provider that exports spans
// over OTLP/HTTP and a W3C trace-context propagator. Tracing activates only when
// an OTLP endpoint is configured through the standard OTEL_EXPORTER_OTLP_ENDPOINT
// (or OTEL_EXPORTER_OTLP_TRACES_ENDPOINT) env var, so local runs stay quiet with
// no collector. The returned func flushes and shuts the provider down — defer it.
func SetupOTel(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	// No collector configured: leave the global no-op provider in place.
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		return noop, nil
	}

	// Default resource (SDK + OTEL_RESOURCE_ATTRIBUTES), with service.name pinned.
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(semconv.ServiceName(serviceName)))
	if err != nil {
		return noop, fmt.Errorf("building otel resource: %w", err)
	}

	// Endpoint, headers, and protocol come from the OTEL_EXPORTER_OTLP_* env vars.
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return noop, fmt.Errorf("creating otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}

// OtelHandler wraps a slog.Handler to inject active OpenTelemetry trace/span IDs
// into structured log records.
type OtelHandler struct {
	slog.Handler
}

// Handle adds tracing context to the slog.Record if a recording span is active.
func (h *OtelHandler) Handle(ctx context.Context, r slog.Record) error {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		sCtx := span.SpanContext()
		if sCtx.HasTraceID() {
			r.AddAttrs(slog.String("trace_id", sCtx.TraceID().String()))
		}
		if sCtx.HasSpanID() {
			r.AddAttrs(slog.String("span_id", sCtx.SpanID().String()))
		}
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new OtelHandler with the given attributes.
func (h *OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OtelHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new OtelHandler with the given group name.
func (h *OtelHandler) WithGroup(name string) slog.Handler {
	return &OtelHandler{Handler: h.Handler.WithGroup(name)}
}
