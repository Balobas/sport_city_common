package outboxRepository

import (
	"context"
	"encoding/json"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
	pgEntity "github.com/balobas/sport_city_common/repository/postgres/entity"
	"github.com/balobas/sport_city_common/tracer"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

func (r *OutboxRepository) UpdateMessage(ctx context.Context, message outboxEntity.Message) error {
	log := repoLoggerFromCtx(ctx).With().Fields(map[string]interface{}{
		"method":  "UpdateMessage",
		"message": message,
	}).Logger()
	log.Debug().Send()

	var span trace.Span = trace.SpanFromContext(ctx)
	traceId, err := getTraceIdFromPayload(message.Payload)
	if err == nil {
		traceId, err := trace.TraceIDFromHex(traceId)
		if err != nil {
			log.Error().Err(err).Msg("invalid trace id in message")
		} else {
			spanContext := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: traceId,
			})
			ctx = trace.ContextWithSpanContext(ctx, spanContext)
		}

		ctx, span = tracer.FromCtx(ctx).Start(ctx, "OutboxRepository.UpdateMessage")
		defer span.End()
	} else {
		log.Warn().Str("error", err.Error()).Msg("failed to get traceId from message payload")
	}

	msgRow := pgEntity.NewOutboxMessageRow().FromEntity(message)

	if err := r.Update(ctx, msgRow, msgRow.ConditionUidEqual()); err != nil {
		err := errors.Wrap(err, "query failed")
		span.RecordError(err)
		log.Debug().Str("error", err.Error()).Send()
		return errors.WithStack(err)
	}
	return nil
}

func getTraceIdFromPayload(payload []byte) (string, error) {
	var traceInfo msgTraceInfo
	if err := json.Unmarshal(payload, &traceInfo); err != nil {
		return "", err
	}
	return traceInfo.TraceId, nil
}

type msgTraceInfo struct {
	TraceId string `json:"traceId"`
}
