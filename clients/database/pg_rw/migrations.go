package pgRw

import (
	"bytes"
	"context"
	"io/fs"

	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/shutdown"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func (c *pgClientRw) Migrate(ctx context.Context, files fs.FS) error {
	db := stdlib.OpenDB(*c.masterPool.Config().ConnConfig)
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
