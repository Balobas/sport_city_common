package useCaseMediaOutbox

import (
	"context"

	"github.com/balobas/sport_city_common/logger"
	"github.com/rs/zerolog"
)

type UcMediaOutbox struct {
	cfg              Config
	outboxRepository OutboxRepository
}

func New(cfg Config, outboxRepo OutboxRepository) *UcMediaOutbox {
	return &UcMediaOutbox{
		cfg:              cfg,
		outboxRepository: outboxRepo,
	}
}

func ucLoggerFromCtx(ctx context.Context) zerolog.Logger {
	log := logger.From(ctx)
	return log.With().Fields(map[string]interface{}{
		"layer":     "usecase",
		"component": "usecaseMediaOutbox",
	}).Logger()
}
