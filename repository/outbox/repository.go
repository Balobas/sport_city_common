package outboxRepository

import (
	"context"

	clientDB "github.com/balobas/sport_city_common/clients/database"
	"github.com/balobas/sport_city_common/logger"
	repositoryBasePostgres "github.com/balobas/sport_city_common/repository/postgres"
	"github.com/rs/zerolog"
)

type OutboxRepository struct {
	*repositoryBasePostgres.BasePgRepository
}

func New(client clientDB.ClientDB) *OutboxRepository {
	return &OutboxRepository{
		repositoryBasePostgres.New(client),
	}
}

func repoLoggerFromCtx(ctx context.Context) zerolog.Logger {
	return logger.From(ctx).With().Fields(map[string]interface{}{
		"layer":     "repository",
		"component": "outboxRepository",
	}).Logger()
}
