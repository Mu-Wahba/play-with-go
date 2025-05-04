package utils

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// initTracer initializes OpenTelemetry with OTLP HTTP exporter
func InitTracer(url string) (*sdktrace.TracerProvider, error) {
	// Create the OTLP HTTP exporter
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(url),
		otlptracehttp.WithInsecure(), // Remove this in production if using TLS
	)
	if err != nil {
		return nil, err
	}

	// Check if exporter is nil
	if exp == nil {
		return nil, fmt.Errorf("failed to create OTLP HTTP exporter: exporter is nil")
	}

	// Create a tracer provider with the OTLP HTTP exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			// semconv.DBSystemPostgreSQL,
			semconv.ServiceName("go-api"),
			// attribute.String("library.language", "go"),
		)),
	)

	// Check if tracer provider is nil
	if tp == nil {
		return nil, fmt.Errorf("failed to create TracerProvider: provider is nil")
	}

	// Set the global tracer provider
	otel.SetTracerProvider(tp)

	// Set the global propagator to tracecontext (for context propagation)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
