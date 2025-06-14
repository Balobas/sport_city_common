package tracer

import (
	"context"
	"encoding/json"

	"github.com/balobas/sport_city_common/logger"
	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func SpanFromCtxWithAttrs(ctx context.Context, methodName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := FromCtx(ctx).Start(ctx, methodName)
	span.SetAttributes(attrs...)
	return ctx, span
}

func UidsToStrSliceAttr(key string, vals []uuid.UUID) attribute.KeyValue {
	strs := make([]string, len(vals))
	for i := 0; i < len(vals); i++ {
		strs[i] = vals[i].String()
	}
	return attribute.StringSlice(key, strs)
}

func SpanParams(ctx context.Context, p map[string]interface{}) attribute.KeyValue {
	log := logger.From(ctx)
	bts, err := json.Marshal(p)
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("failed to marshal params for trace span")
		return attribute.KeyValue{}
	}
	return attribute.String("params", string(bts))
}
