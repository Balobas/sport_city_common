package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func SpanFromCtxWithAttrs(ctx context.Context, methodName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := FromCtx(ctx).Start(ctx, methodName)
	span.SetAttributes(attrs...)
	return ctx, span
}

func StringerToStrSliceAttr(key string, vals []fmt.Stringer) attribute.KeyValue {
	strs := make([]string, len(vals))
	for i := 0; i < len(vals); i++ {
		strs[i] = vals[i].String()
	}
	return attribute.StringSlice(key, strs)
}
