package loggingInterceptor

import (
	"context"
	"errors"
	"time"

	"github.com/balobas/sport_city_common/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func UnaryLoggingInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	var op options

	for _, opt := range opts {
		opt(&op)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log := logger.Logger()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Info().Str("req", info.FullMethod).Msg("metadata not provided")
			return handler(ctx, req)
		}

		var reqId string

		reqIdMd := md.Get("reqId")
		if len(reqIdMd) != 0 && len(reqIdMd[0]) != 0 {
			reqId = reqIdMd[0]
		}
		log = log.With().Str("req_id", reqId).Logger()

		ctx = logger.ContextWithLogger(ctx, log)

		logRequestFields := map[string]interface{}{
			"method": info.FullMethod,
		}

		p, ok := peer.FromContext(ctx)
		if ok {
			logRequestFields["remote_ip"] = p.Addr.String()
		}

		if op.withRequest {
			var requestBody string

			if msg, ok := req.(proto.Message); ok {
				bts, err := protojson.Marshal(msg)
				if err != nil {
					log.Debug().Err(err).Msg("failed to marshall request to json")
					requestBody = "cant unmarshall into json"
				} else {
					requestBody = string(bts)
				}
			} else {
				requestBody = "cant assert to proto.Message"
			}

			logRequestFields["request_body"] = requestBody
		}

		log.Info().
			Timestamp().
			Fields(logRequestFields).
			Msg("incoming request")
		t1 := time.Now()

		resp, err := handler(ctx, req)

		t2 := time.Now()

		logResponseFields := map[string]interface{}{
			"method":     info.FullMethod,
			"status":     status.Code(err).String(),
			"latency_ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
		}
		if err != nil {
			logResponseFields["error"] = err.Error()

			if op.shouldLogError(err) {
				log.Error().Err(err).Send()
			}
		}

		if op.withResponse {
			var respBody string
			if msg, ok := resp.(proto.Message); ok {
				bts, err := protojson.Marshal(msg)
				if err != nil {
					log.Debug().Err(err).Msg("failed to marshall response to json")
					respBody = "cant unmarshall into json"
				} else {
					respBody = string(bts)
				}
			} else {
				respBody = "cant assert to proto message"
			}

			logResponseFields["response"] = respBody
		}

		// log end request
		log.Info().
			Timestamp().
			Fields(logResponseFields).
			Msg("incoming request finished")

		return resp, err
	}
}

type options struct {
	withRequest      bool
	withResponse     bool
	withoutLogErrors []error
}

type Option func(opts *options)

func WithRequestLog() Option {
	return func(opts *options) {
		opts.withRequest = true
	}
}

func WithResponseLog() Option {
	return func(opts *options) {
		opts.withResponse = true
	}
}

func WithAllLog() Option {
	return func(opts *options) {
		opts.withRequest = true
		opts.withResponse = true
	}
}

func WithoutLogErrors(errs ...error) Option {
	return func(opts *options) {
		opts.withoutLogErrors = errs
	}
}

func (opts *options) shouldLogError(err error) bool {
	for _, e := range opts.withoutLogErrors {
		if errors.Is(err, e) {
			return false
		}
	}
	return true
}
