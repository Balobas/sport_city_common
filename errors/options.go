package commonErrors

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type errorOptions struct {
	withSpanRecord bool
	spanForRecord  trace.Span

	withDebugLog   bool
	loggerForDebug *zerolog.Logger
}

type ErrorOption func(*errorOptions)

func WithSpanRecord(span trace.Span) ErrorOption {
	return func(eo *errorOptions) {
		eo.withSpanRecord = true
		eo.spanForRecord = span
	}
}

func WithDebugLog(log *zerolog.Logger) ErrorOption {
	return func(eo *errorOptions) {
		eo.withDebugLog = true
		eo.loggerForDebug = log
	}
}
