package apm

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel/attribute"
)

var (
	once      sync.Once
	activeAPM *APM
)

// list supported APM
const (
	DatadogAPMType = iota
	OpenTelemetryAPMType
)

type APMPayload struct {
	// ServiceName, ServiceEnv, ServiceTribe, ServiceTribe is required
	ServiceName    string
	ServiceEnv     string
	ServiceVersion string
	ServiceTribe   string // tribe of service

	// ServiceHost is optional
	ServiceHost *string
	SampleRate  *float64
}

// tracer interface
type Tracer interface {
	EndAPM()
	StartTransaction(ctx context.Context, name string) (context.Context, interface{})
	EndTransaction(txn interface{})
	GetTraceID(ctx context.Context) string
	AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)
}

type APM struct {
	tracer Tracer
}

func NewAPM(apmType int, payload APMPayload) (*APM, error) {
	// validate payload for required field
	if payload.ServiceName == "" {
		return nil, errors.New("ServiceName is required")
	}
	if payload.ServiceEnv == "" {
		return nil, errors.New("ServiceEnv is required")
	}
	if payload.ServiceVersion == "" {
		return nil, errors.New("ServiceVersion is required")
	}
	if payload.ServiceTribe == "" {
		return nil, errors.New("ServiceTribe is required")
	}

	// validate payload for required field based on apmType

	// create tracer
	var tracer Tracer
	var err error
	switch apmType {
	case DatadogAPMType:
		tracer, err = NewDatadogTracer(payload.ServiceHost, payload.ServiceName, payload.ServiceEnv, payload.ServiceVersion, payload.ServiceTribe, payload.SampleRate)
	case OpenTelemetryAPMType:
		tracer, err = NewOpenTelemetryTracer(payload.ServiceHost, payload.ServiceName, payload.ServiceEnv, payload.ServiceTribe, payload.SampleRate)

	default:
		return nil, errors.New("unsupported APM type")
	}
	if err != nil {
		return nil, err
	}

	APM := &APM{tracer: tracer}
	// set global APM
	once.Do(func() { activeAPM = APM })

	return APM, nil
}

func StartTransaction(ctx context.Context, name string) (context.Context, interface{}) {
	if activeAPM == nil {
		return ctx, nil
	}
	return activeAPM.tracer.StartTransaction(ctx, name)
}

func EndTransaction(txn interface{}) {
	if activeAPM != nil {
		activeAPM.tracer.EndTransaction(txn)
	}
}

func (a *APM) EndAPM() {
	if a.tracer != nil {
		a.tracer.EndAPM()
	}
}

func GetTraceID(ctx context.Context) string {
	if activeAPM == nil {
		return ""
	}
	return activeAPM.tracer.GetTraceID(ctx)
}

func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if activeAPM != nil {
		activeAPM.tracer.AddEvent(ctx, name, attrs...)
	}
}
