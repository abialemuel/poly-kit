package apm

import (
	"context"
	"fmt"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DatadogTracer struct{}

var (
	defaultSampleRate = 1.0
)

func NewDatadogTracer(serviceHost *string, serviceName, serviceEnv, serviceVersion, serviceTribe string, sampleRate *float64) (*DatadogTracer, error) {
	// Create a new Datadog tracer with the given  service name, env and version.
	defaultServiceHost := "localhost:8126"
	if serviceHost == nil {
		serviceHost = &defaultServiceHost
	}

	if sampleRate == nil {
		sampleRate = &defaultSampleRate
	}

	tracer.Start(
		tracer.WithEnv(serviceEnv),
		tracer.WithAgentAddr(*serviceHost),
		tracer.WithService(serviceName),
		tracer.WithServiceVersion(serviceVersion),
		tracer.WithGlobalTag("tribe", serviceTribe),
		tracer.WithGlobalTag("platform", "go"),
		tracer.WithSampler(tracer.NewRateSampler(*sampleRate)),
	)

	return &DatadogTracer{}, nil
}

func (t *DatadogTracer) StartTransaction(ctx context.Context, name string) (context.Context, interface{}) {
	// Start a new span with the given name and options.
	span, ctx := tracer.StartSpanFromContext(ctx, name)
	return ctx, span
}

func (t *DatadogTracer) EndTransaction(txn interface{}) {
	// End the given span.
	span := txn.(tracer.Span)
	span.Finish()
}

func (t *DatadogTracer) EndAPM() {
	tracer.Stop()
}

func (t *DatadogTracer) GetTraceID(ctx context.Context) string {
	// Get the trace ID from the given context.
	span, _ := tracer.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return strconv.FormatUint(span.Context().TraceID(), 10)
}

func (t *DatadogTracer) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	// Add an event to the span.
	span, ok := tracer.SpanFromContext(ctx)
	if !ok || span == nil {
		return
	}

	for _, attr := range attrs {
		key := fmt.Sprintf("%s.%s", name, string(attr.Key))
		value := attr.Value.AsInterface()

		// Validate key and value
		if key == "" || value == nil {
			continue // Skip invalid attributes
		}

		// Special handling for "error" attributes
		if attr.Key == attribute.Key("error") {
			span.SetTag("error", attr.Value) // Use string representation
		} else {
			// Safely set other tags
			span.SetTag(key, value)
		}
	}

}
