package entrypoint

import (
	"context"

	entrypointGrpc "github.com/balobas/sport_city_common/entrypoint/grpc"
	entrypointHttp "github.com/balobas/sport_city_common/entrypoint/http"
	entrypointProccess "github.com/balobas/sport_city_common/entrypoint/proccess"
	"github.com/balobas/sport_city_common/tracer"
)

type Option func(*EntryPoint)

func (ep *EntryPoint) applyOpts(opts ...Option) {
	for _, opt := range opts {
		opt(ep)
	}
}

func WithTracing(cfg tracer.Config) Option {
	return func(ep *EntryPoint) {
		pr := make([]PreRunFn, 0, len(ep.preRuns)+1)
		pr = append(pr, func(ctx context.Context) error {
			return initTracer(cfg)
		})
		ep.preRuns = append(pr, ep.preRuns...)
	}
}

func WithPreRuns(preRuns ...PreRunFn) Option {
	return func(ep *EntryPoint) {
		ep.preRuns = append(ep.preRuns, preRuns...)
	}
}

func WithProcesses(processes ...entrypointProccess.Process) Option {
	return func(ep *EntryPoint) {
		ep.processes = processes
	}
}

func WithGrpcServers(srvs ...*entrypointGrpc.Server) Option {
	return func(ep *EntryPoint) {
		ep.grpcServers = srvs
	}
}

func WithHttpServers(srvs ...*entrypointHttp.Server) Option {
	return func(ep *EntryPoint) {
		ep.httpServers = srvs
	}
}
