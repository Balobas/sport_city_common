package entrypointProccess

import (
	"context"

	"github.com/balobas/sport_city_common/logger"
)

func Run(ctx context.Context, process Process, done chan<- struct{}) {
	log := logger.From(ctx)

	go func() {
		processErr := make(chan error, 1)

		go func() {
			processErr <- process.Run(ctx)
			close(processErr)
		}()

		select {
		case <-ctx.Done():
		case err := <-processErr:
			if err != nil {
				log.Error().Err(err).Msgf("failed to run process %s", process.Name())
				done <- struct{}{}
			} else {
				log.Info().Msgf("process %s start running", process.Name())
			}
		}
		return
	}()
}
