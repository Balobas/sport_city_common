package pgRw

import (
	"context"

	common "github.com/balobas/sport_city_common"
	"github.com/balobas/sport_city_common/logger"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (p *pgClientRw) BeginTxWithContext(ctx context.Context, isolationLevel string) (context.Context, common.Transaction, error) {
	log := logger.From(ctx)

	if tx, ok := p.GetTxFromCtx(ctx); ok {
		log.Debug().Msg("pg: tx already exist in ctx")
		return ctx, tx, nil
	}

	tx, err := p.getConnByCtxKey(ctx).BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.TxIsoLevel(isolationLevel)})
	if err != nil {
		log.Debug().Msg("failed to begin tx")
		return ctx, nil, errors.Wrap(err, "failed to begin tx")
	}

	log.Debug().Msg("begin new tx")
	return context.WithValue(ctx, TxKey{}, tx), tx, nil
}

func (p *pgClientRw) HasTxInCtx(ctx context.Context) bool {
	_, ok := ctx.Value(TxKey{}).(pgx.Tx)
	return ok
}

func (p *pgClientRw) GetTxFromCtx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(TxKey{}).(pgx.Tx)
	return tx, ok
}
