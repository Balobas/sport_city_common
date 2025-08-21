package pg

import (
	"context"

	clientDB "github.com/balobas/sport_city_common/clients/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type pgClient struct {
	dbc *pg
}

func NewClient(ctx context.Context, dsn string) (clientDB.ClientDB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, errors.Errorf("failed to ping db: %v", err)
	}

	return &pgClient{
		dbc: &pg{pool: pool},
	}, nil
}

func (c *pgClient) DB() clientDB.DB {
	return c.dbc
}

func (c *pgClient) Close(ctx context.Context) error {
	if c.dbc != nil {
		c.dbc.Close()
	}

	return nil
}

func (c *pgClient) GetPool() *pgxpool.Pool {
	return c.dbc.pool
}
