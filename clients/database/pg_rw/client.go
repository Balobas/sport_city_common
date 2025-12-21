package pgRw

import (
	"context"
	"strings"

	clientDB "github.com/balobas/sport_city_common/clients/database"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type pgClientRw struct {
	masterPool  *pgxpool.Pool
	replicaPool *pgxpool.Pool
}

func NewClientRW(
	ctx context.Context,
	masterDsn string, masterOpts []PgClientOption,
	replicaDsn string, replicaOpts []PgClientOption,
) (clientDB.ClientDB, error) {
	masterPool, err := newPool(ctx, masterDsn, masterOpts...)
	if err != nil {
		return nil, err
	}

	replicaPool, err := newPool(ctx, replicaDsn, append(replicaOpts, withReadOnly())...)
	if err != nil {
		return nil, err
	}

	return &pgClientRw{
		replicaPool: masterPool,
		masterPool:  replicaPool,
	}, nil
}

func newPool(ctx context.Context, dsn string, opts ...PgClientOption) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Errorf("failed to parse dsn to config: %v", err)
	}
	options := &pgClientOptions{}
	for _, applyOpt := range opts {
		applyOpt(options)
	}

	if options.readOnly {
		cfg.AfterConnect = setReadOnly
	}
	if options.maxConns > 0 {
		cfg.MaxConns = options.maxConns
	}
	if options.minConns > 0 {
		cfg.MinConns = options.minConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, errors.Errorf("failed to ping db: %v", err)
	}

	return pool, nil
}

func (p *pgClientRw) Close(ctx context.Context) error {
	p.replicaPool.Close()
	p.masterPool.Close()
	return nil
}

func (c *pgClientRw) GetMasterPool() *pgxpool.Pool {
	return c.masterPool
}

func (c *pgClientRw) GetReplicaPool() *pgxpool.Pool {
	return c.replicaPool
}

func (p *pgClientRw) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	tx, ok := p.GetTxFromCtx(ctx)
	if ok {
		tag, err := tx.Exec(ctx, sql, args...)
		return tag, convertError(err)
	}

	tag, err := p.getConnByCtxKey(ctx).Exec(ctx, sql, args...)
	return tag, convertError(err)
}

func (p *pgClientRw) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	tx, ok := p.GetTxFromCtx(ctx)
	if ok {
		rows, err := tx.Query(ctx, sql, args...)
		return rows, convertError(err)
	}

	rows, err := p.getConnByCtxKey(ctx).Query(ctx, sql, args...)
	return rows, convertError(err)
}

func (p *pgClientRw) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	tx, ok := p.GetTxFromCtx(ctx)
	if ok {
		return tx.QueryRow(ctx, sql, args...)
	}

	return p.getConnByCtxKey(ctx).QueryRow(ctx, sql, args...)
}

func (p *pgClientRw) ScanQueryRow(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	row, err := p.Query(ctx, sql, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (p *pgClientRw) ScanAllQuery(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	rows, err := p.Query(ctx, sql, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (p *pgClientRw) Ping(ctx context.Context) error {
	errs := make(map[string]error)
	err := p.masterPool.Ping(ctx)
	if err != nil {
		errs["master_pool"] = err
	}

	err = p.replicaPool.Ping(ctx)
	if err != nil {
		errs["replica_pool"] = err
	}

	if len(errs) == 0 {
		return nil
	}

	errMsg := strings.Builder{}
	for k, v := range errs {
		errMsg.WriteString(k)
		errMsg.WriteString(" has error")
		errMsg.WriteString(v.Error())
		errMsg.WriteByte(',')
	}

	return errors.New(errMsg.String())
}
