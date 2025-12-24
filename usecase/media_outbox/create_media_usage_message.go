package useCaseMediaOutbox

import (
	"context"
	"encoding/json"
	"time"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
	"github.com/balobas/sport_city_common/tracer"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/trace"
)

// Deprecated
func (uc *UcMediaOutbox) CreateMediaUsageMessage(
	ctx context.Context,
	domain string,
	firstlyUsedMedia []uuid.UUID,
	unusedMedia []uuid.UUID,
) error {
	ctx, span := tracer.FromCtx(ctx).Start(ctx, "UseCaseMediaOutbox.CreateMediaUsageMessage")
	defer span.End()

	log := ucLoggerFromCtx(ctx).With().Fields(map[string]interface{}{
		"method":           "CreateMediaUsageMessage",
		"firstlyUsedMedia": firstlyUsedMedia,
		"unusedMedia":      unusedMedia,
	}).Logger()
	log.Debug().Send()

	if len(firstlyUsedMedia) == 0 && len(unusedMedia) == 0 {
		log.Debug().Msg("empty firstlyUsedMedia and unused media. ignore")
		return nil
	}

	msgPayload := outboxEntity.MediaUsagePayload{
		Domain:      domain,
		FirstlyUsed: firstlyUsedMedia,
		Unused:      unusedMedia,
	}
	msgPayload.MsgUid = uuid.NewV4()
	msgPayload.TraceId = span.SpanContext().TraceID().String()

	payloadBts, err := json.Marshal(msgPayload)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		log.Debug().Str("error", err.Error()).Msg("failed to marshal msg payload")
		return errors.WithStack(err)
	}

	msg := outboxEntity.Message{
		Uid:         msgPayload.MsgUid,
		SubjectName: uc.cfg.MediaUsageMessageSubject(),
		Payload:     payloadBts,
		CreatedAt:   time.Now().UTC(),
	}

	return uc.outboxRepository.CreateMessage(ctx, msg)
}
