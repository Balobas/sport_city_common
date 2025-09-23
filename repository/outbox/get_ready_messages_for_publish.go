package outboxRepository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
	pgEntity "github.com/balobas/sport_city_common/repository/postgres/entity"
	"github.com/pkg/errors"
)

func (r *OutboxRepository) GetReadyMessagesForPublish(ctx context.Context, batchSize int64) ([]outboxEntity.Message, error) {
	log := repoLoggerFromCtx(ctx).With().Str("method", "GetReadyMessagesForPublish").Logger()
	// log.Debug().Msgf("outboxRepository.GetReadyMessagesForPublish: batch size %d", batchSize)

	msgRow := pgEntity.NewOutboxMessageRow()

	sql, args, err := sq.Select(
		msgRow.Columns()...,
	).From(
		msgRow.Table(),
	).PlaceholderFormat(
		sq.Dollar,
	).Where(
		msgRow.ConditionSendAtIsNull(),
	).OrderBy("created_at").Limit(uint64(batchSize)).ToSql()
	if err != nil {
		err = errors.Wrap(err, "failed to build sql for GetReadyMessagesForPublish")
		log.Debug().Str("error", err.Error()).Send()
		return nil, errors.WithStack(err)
	}

	rows, err := r.DB().Query(ctx, sql, args...)
	if err != nil {
		err = errors.Wrap(err, "query failed")
		log.Debug().Str("error", err.Error()).Send()
		return nil, errors.WithStack(err)
	}

	msgRows := pgEntity.NewOutboxMessageRows()
	if err := msgRows.ScanAll(rows); err != nil {
		err = errors.Wrap(err, "scan failed")
		log.Debug().Str("error", err.Error()).Send()
		return nil, errors.WithStack(err)
	}

	return msgRows.ToEntity(), nil
}
