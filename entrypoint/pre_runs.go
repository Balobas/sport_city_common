package entrypoint

import (
	"context"

	"github.com/balobas/sport_city_common/tracer"
)

type PreRunFn func(context.Context) error

func initTracer(cfg tracer.Config) error {
	_, err := tracer.NewTracerProvider(cfg)
	return err
}
