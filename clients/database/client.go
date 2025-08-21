package DBclient

import (
	"context"

	common "github.com/balobas/sport_city_common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryExecer interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	ScanQueryRow(ctx context.Context, dest interface{}, sql string, args ...interface{}) error
	ScanAllQuery(ctx context.Context, dest interface{}, sql string, args ...interface{}) error
}

type Transactor interface {
	BeginTxWithContext(ctx context.Context, isolationLevel string) (context.Context, common.Transaction, error)
}

type DB interface {
	QueryExecer
	Transactor
	HasTxInCtx(ctx context.Context) bool
	Ping(ctx context.Context) error
	Close()
}

type ClientDB interface {
	DB() DB
	GetPool() *pgxpool.Pool
	Close(ctx context.Context) error
}
