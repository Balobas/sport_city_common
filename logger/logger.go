package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	zLog "github.com/rs/zerolog/log"
)

type Config interface {
	Debug() bool
}

const TimeFormat = "2006-01-02T15:04:05.000000Z07:00"

func Init(cfg Config) {
	zerolog.TimeFieldFormat = TimeFormat
	zerolog.ErrorStackMarshaler = MarshalMultiStack

	zL := zerolog.New(os.Stdout).
		With().
		Stack().
		Caller().
		Logger()

	if cfg.Debug() {
		zL = zL.Level(zerolog.DebugLevel)
	} else {
		zL = zL.Level(zerolog.InfoLevel)
	}

	zLog.Logger = zL
}

func ContextWithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

func From(ctx context.Context) zerolog.Logger {
	return *zLog.Ctx(ctx)
}

func Logger() zerolog.Logger {
	return zLog.Logger
}
