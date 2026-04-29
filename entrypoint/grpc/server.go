package entrypointGrpc

import (
	"context"
	"net"

	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	authInterceptor "github.com/balobas/sport_city_common/grpc/interceptors/auth"
	loggingInterceptor "github.com/balobas/sport_city_common/grpc/interceptors/logging"
	tracingInterceptor "github.com/balobas/sport_city_common/grpc/interceptors/tracing"
)

type Server struct {
	srv     GrpcService
	cfg     Config
	srvOpts *ServerOptions
}

func NewServer(
	cfg Config,
	srv GrpcService,
	opts ...ServerOption,
) *Server {
	srvOpts := NewServerOptions()
	srvOpts.Apply(opts...)

	return &Server{
		cfg:     cfg,
		srv:     srv,
		srvOpts: srvOpts,
	}
}

func (s *Server) Run(
	ctx context.Context,
	done chan<- struct{},
) {
	grpcServer := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.ChainUnaryInterceptor(s.buildInterceptors(s.srvOpts)...),
	)

	if !s.srvOpts.WithoutReflection() {
		reflection.Register(grpcServer)
	}
	s.srv.Register(grpcServer)

	log := logger.From(ctx)

	log.Info().Msgf("grpc server is running on %s", s.cfg.ServiceGrpcAddress())

	go func() {
		lis, err := net.Listen("tcp", s.cfg.ServiceGrpcAddress())
		if err != nil {
			log.Error().Err(err).Msg("failed to listen tcp")
			done <- struct{}{}
			return
		}

		shutdown.Add(func(ctx context.Context) error {
			grpcServer.Stop()
			return nil
		})

		serverStopped := make(chan struct{})
		go func() {
			err := grpcServer.Serve(lis)
			if err != nil {
				log.Error().Err(err).Msg("grpc server cancelled with error")
			} else {
				log.Info().Msg("grpc server cancelled without errors")
			}
			close(serverStopped)
		}()

		select {
		case <-ctx.Done():
			log.Info().Msgf("grpc server cancelled, ctx.Done, error: %v", ctx.Err())
			return
		case <-serverStopped:
			log.Info().Msg("grpc server cancelled")
			done <- struct{}{}
		}
	}()
}

func (s *Server) buildInterceptors(opts *ServerOptions) []grpc.UnaryServerInterceptor {
	interceptors := []grpc.UnaryServerInterceptor{}

	if opts.WithLoggingInterceptor() {
		interceptors = append(interceptors, loggingInterceptor.UnaryLoggingInterceptor(opts.LoggingInterceptorOptions()...))
	}
	if opts.WithTracingInterceptor() {
		interceptors = append(interceptors, tracingInterceptor.UnaryTracingInterceptor())
	}
	if opts.WithAuthInterceptor() {
		authOpts := opts.AuthInterceptorOptions()
		interceptors = append(interceptors, authInterceptor.UnaryAuthInterceptor(authOpts.WithoutAuthMap()))
	}
	if opts.WithCustomInterceptors() {
		interceptors = append(interceptors, opts.CustomInterceptors()...)
	}
	return interceptors
}
