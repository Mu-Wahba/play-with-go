package utils

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func InitGlobalTracer(tp trace.TracerProvider) {
	Tracer = tp.Tracer("event-service")
}

func DbQueryAttributes(query string, rowsAffected int, duration time.Duration, operation string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.statement", query),
		attribute.Int("db.rows_affected", rowsAffected),
		attribute.String("db.operation", operation),
		attribute.String("db.type", "sql"),
		attribute.Float64("db.query_duration_seconds", duration.Seconds()), // Duration in seconds
	}
}

// setErrorOnSpan sets error information on the given span.
func SetErrorOnSpan(span trace.Span, err error) {

	// Set the status of the span to 'Error'
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	// Set additional error-related attributes
	span.SetAttributes(
		attribute.String("error.message", err.Error()),
	)
}

// setErrorOnSpan sets error information on the given span.
func SetOkMsgOnSpan(span trace.Span, code codes.Code, description string, msg string) {

	// Set the status of the span to 'Error'
	span.SetStatus(codes.Ok, description)

	// Set additional error-related attributes
	span.SetAttributes(
		attribute.String("message", msg),
	)
}
