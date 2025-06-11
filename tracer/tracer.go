package tracer

import (
	"context"
	"fmt"
	"sync"

	jaegerExporter "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	defaultTracer trace.Tracer
	once          = sync.Once{}
)

type Config interface {
	ServiceName() string
	TraceCollectorUrl() string
}

func NewTracerProvider(cfg Config) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaegerExporter.New(
		jaegerExporter.WithCollectorEndpoint(
			jaegerExporter.WithEndpoint(
				cfg.TraceCollectorUrl(),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName()),
		)),
	)
	once.Do(func() {
		defaultTracer = tp.Tracer(fmt.Sprintf("%s_tracer", cfg.ServiceName()))
	})

	return tp, nil
}

type tracerCtxKey struct{}

func ContextWithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerCtxKey{}, tracer)
}

func FromCtx(ctx context.Context) trace.Tracer {
	if tracer, ok := ctx.Value(tracerCtxKey{}).(trace.Tracer); ok {
		return tracer
	}
	return defaultTracer
}
