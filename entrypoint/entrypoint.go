package entrypoint

import (
	"context"
	"time"

	entrypointGrpc "github.com/balobas/sport_city_common/entrypoint/grpc"
	entrypointHttp "github.com/balobas/sport_city_common/entrypoint/http"
	entrypointProccess "github.com/balobas/sport_city_common/entrypoint/proccess"
	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/shutdown"
)

type EntryPoint struct {
	cfg       Config
	preRuns   []PreRunFn
	processes []entrypointProccess.Process

	grpcServers []*entrypointGrpc.Server
	httpServers []*entrypointHttp.Server
}

func New(
	cfg Config,
	opts ...Option,
) *EntryPoint {
	ep := &EntryPoint{}
	ep.applyOpts(opts...)
	return ep
}

func (ep *EntryPoint) Run(ctx context.Context) {
	done := make(chan struct{}, 10)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger.Init(ep.cfg)

	log := logger.Logger()
	ctx = logger.ContextWithLogger(ctx, log)

	defer func() {
		if r := recover(); r != nil {
			l := log.Error()
			if err, ok := r.(error); ok {
				l = l.Err(err)
			}
			l.Msgf("recovered panic: error %s", r)
		}

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancelShutdown()

		log.Info().Msg("start shutdown")
		shutdown.CloseAll(shutdownCtx)
		close(done)
	}()

	for _, preRun := range ep.preRuns {
		if err := preRun(ctx); err != nil {
			log.Error().Err(err).Msg("failed to preRun")
			return
		}
	}

	for _, process := range ep.processes {
		entrypointProccess.Run(ctx, process, done)
	}

	for _, grpcSrv := range ep.grpcServers {
		grpcSrv.Run(ctx, done)
	}

	for _, httpSrv := range ep.httpServers {
		httpSrv.Run(ctx, done)
	}

	log.Info().Msg("service started")
	select {
	case <-ctx.Done():
		log.Info().Msgf("ctx.Done: %v", ctx.Err())
	case <-done:
		log.Warn().Msg("one of the critical components closed")
	}
}
