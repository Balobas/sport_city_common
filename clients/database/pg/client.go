package pg

import (
	"bytes"
	"context"
	"io/fs"

	clientDB "github.com/balobas/sport_city_common/clients/database"
	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/shutdown"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

type pgClient struct {
	*pg
	name string
	cfg  *pgxpool.Config
}

func NewClient(ctx context.Context, name string, dsn string, opts ...PgClientOption) (clientDB.ClientDB, error) {
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

	return &pgClient{
		pg:   &pg{pool: pool},
		name: name,
		cfg:  cfg,
	}, nil
}

func (c *pgClient) Name() string {
	return c.name
}

func (c *pgClient) Close(ctx context.Context) error {
	if c.pg != nil {
		c.pg.Close()
	}
	return nil
}

func (c *pgClient) GetMasterPool() *pgxpool.Pool {
	return c.pg.pool
}

func (c *pgClient) Migrate(ctx context.Context, files fs.FS) error {
	db := stdlib.OpenDB(*c.cfg.ConnConfig)
	shutdown.Add(func(ctx context.Context) error {
		return db.Close()
	})

	pgLock, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return errors.Errorf("failed to create pg lock: %v", err)
	}

	fileSys, err := fs.Sub(files, "sql")
	if err != nil {
		return errors.Errorf("failed to create fs: %v", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres, db, fileSys,
		goose.WithSessionLocker(pgLock),
	)
	if err != nil {
		return errors.Errorf("failed to create provider: %v", err)
	}

	res, err := provider.Up(ctx)
	if err != nil {
		return errors.Errorf("failed to migrate: %v", err)
	}

	log := logger.From(ctx)
	log.Info().Str("result", convertMigrationResultToStr(res)).Msg("database migrated successfully")
	return nil
}

func convertMigrationResultToStr(res []*goose.MigrationResult) string {
	var buf bytes.Buffer

	for _, r := range res {
		buf.WriteString("\n" + r.String())
	}

	if len(buf.Bytes()) == 0 {
		buf.WriteString("No migrations to apply")
	}

	return buf.String()
}

func (c *pgClient) CtxWithMasterKey(ctx context.Context) context.Context {
	return ctx
}

func (c *pgClient) CtxWithReplicaKey(ctx context.Context) context.Context {
	return ctx
}
