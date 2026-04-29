package entrypointHttp

import (
	"context"
	"net/http"

	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/shutdown"
)

type Server struct {
	cfg     Config
	handler http.Handler
}

func NewServer(cfg Config, handler http.Handler) *Server {
	return &Server{
		cfg:     cfg,
		handler: handler,
	}
}

func (s *Server) Run(ctx context.Context, done chan<- struct{}) {
	log := logger.From(ctx)

	server := http.Server{
		Addr:    s.cfg.HttpHost() + ":" + s.cfg.HttpPort(),
		Handler: s.handler,
	}
	shutdown.Add(server.Shutdown)

	log.Info().Msgf("http server running on: %s", s.cfg.HttpHost()+":"+s.cfg.HttpPort())

	go func() {

		stopped := make(chan struct{}, 1)

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Info().Err(err).Msg("http server stopped with error")
			} else {
				log.Info().Msg("http server stopped")
			}
			close(stopped)
		}()

		select {
		case <-ctx.Done():
		case <-stopped:
			done <- struct{}{}
		}
	}()
}
