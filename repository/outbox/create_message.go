package outboxRepository

import (
	"context"

	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
	pgEntity "github.com/balobas/sport_city_common/repository/postgres/entity"
	"github.com/balobas/sport_city_common/tracer"
	"github.com/pkg/errors"
)

func (r *OutboxRepository) CreateMessage(ctx context.Context, message outboxEntity.Message) error {
	ctx, span := tracer.FromCtx(ctx).Start(ctx, "OutboxRepository.CreateMessage")
	defer span.End()

	log := repoLoggerFromCtx(ctx).With().Fields(map[string]interface{}{
		"method":  "CreateMessage",
		"message": message,
	}).Logger()
	log.Debug().Send()

	if err := r.Create(ctx, pgEntity.NewOutboxMessageRow().FromEntity(message)); err != nil {
		err = errors.Wrap(err, "query failed")
		span.RecordError(err)
		log.Debug().Str("error", err.Error()).Send()
		return errors.WithStack(err)
	}
	return nil
}
