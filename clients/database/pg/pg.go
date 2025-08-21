package pg

import (
	"context"

	common "github.com/balobas/sport_city_common"
	"github.com/balobas/sport_city_common/logger"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type TxKey struct{}

type pg struct {
	pool *pgxpool.Pool
}

type (
	ExecFn  func(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryFn func(context.Context, string, ...any) (pgx.Rows, error)
)

func (p *pg) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	var execFn ExecFn
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		execFn = tx.Exec
	} else {
		execFn = p.pool.Exec
	}

	tag, err := execFn(ctx, sql, args...)
	return tag, convertError(err)
}

func (p *pg) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	var queryFn QueryFn
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		queryFn = tx.Query
	} else {
		queryFn = p.pool.Query
	}

	rows, err := queryFn(ctx, sql, args...)
	return rows, convertError(err)
}

func (p *pg) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}

	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *pg) ScanQueryRow(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	row, err := p.Query(ctx, sql, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (p *pg) ScanAllQuery(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	rows, err := p.Query(ctx, sql, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (p *pg) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *pg) Close() {
	p.pool.Close()
}

func (p *pg) BeginTxWithContext(ctx context.Context, isolationLevel string) (context.Context, common.Transaction, error) {
	log := logger.From(ctx)

	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		log.Debug().Msg("pg: tx already exist in ctx")
		return ctx, tx, nil
	}

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.TxIsoLevel(isolationLevel)})
	if err != nil {
		log.Debug().Msg("failed to begin tx")
		return ctx, nil, errors.Wrap(err, "failed to begin tx")
	}

	log.Debug().Msg("begin new tx")
	return context.WithValue(ctx, TxKey{}, tx), tx, nil
}

func (p *pg) HasTxInCtx(ctx context.Context) bool {
	_, ok := ctx.Value(TxKey{}).(pgx.Tx)
	return ok
}
