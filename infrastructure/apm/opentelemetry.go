package apm

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semConv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OpenTelemetryTracer struct {
	tracer trace.Tracer
}

func NewOpenTelemetryTracer(serviceHost *string, serviceName string, serviceEnv, serviceTribe string, sampleRate *float64) (*OpenTelemetryTracer, error) {
	var (
		err               error
		conn              *grpc.ClientConn
		defaultSampleRate = 1.0
	)
	conn, err = grpc.DialContext(context.Background(), *serviceHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		return nil, err
	}

	// Create a new OTLP exporter over gRPC
	exp, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	traceExporter := sdkTrace.SpanExporter(exp)
	// Create a new trace provider with the exporter.

	if sampleRate == nil {
		sampleRate = &defaultSampleRate
	}
	tp := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(traceExporter),
		sdkTrace.WithResource(resource.NewWithAttributes(
			semConv.SchemaURL,
			semConv.ServiceNameKey.String(serviceName),
			attribute.String("tribe", serviceTribe),
			attribute.String("env", serviceEnv),
			attribute.String("version", "ver.1"),
			attribute.String("platform", "go"),
		)),
		sdkTrace.WithSampler(sdkTrace.TraceIDRatioBased(*sampleRate)),
	)

	// Set the global trace provider and the propagation.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{}, propagation.TraceContext{}))

	tracer := otel.GetTracerProvider().Tracer(serviceName)

	return &OpenTelemetryTracer{tracer: tracer}, nil
}

func (o *OpenTelemetryTracer) StartTransaction(ctx context.Context, name string) (context.Context, interface{}) {
	// Start a new OpenTelemetry span with the given name from a context.
	ctx, span := o.tracer.Start(ctx, name)
	return ctx, span
}

func (o *OpenTelemetryTracer) EndTransaction(txn interface{}) {
	// End the given OpenTelemetry span.
	span := txn.(trace.Span)
	span.End()
}

func (o *OpenTelemetryTracer) EndAPM() {
	// shutdown the tracer
	if tp, ok := otel.GetTracerProvider().(*sdkTrace.TracerProvider); ok {
		tp.Shutdown(context.Background())
	}
}

// get trace id
func (o *OpenTelemetryTracer) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	sc := span.SpanContext()
	return sc.TraceID().String()
}

// Add event log to span
func (o *OpenTelemetryTracer) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.AddEvent(name, trace.WithAttributes(attrs...))

	// if attrs contains key error, set the span status to error
	for _, attr := range attrs {
		if attr.Key == attribute.Key("error") {
			span.SetStatus(codes.Error, attr.Value.Emit())
		}
	}
}
