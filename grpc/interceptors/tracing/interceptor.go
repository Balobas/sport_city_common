package tracingInterceptor

import (
	"context"
	"fmt"

	"github.com/balobas/sport_city_common/logger"
	"github.com/balobas/sport_city_common/tracer"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const traceIdHeader = "x-trace-id"

func UnaryTracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log := logger.From(ctx).With().Str("component", "tracingInterceptor").Logger()

		md, _ := metadata.FromIncomingContext(ctx)
		if len(md[traceIdHeader]) != 0 {
			traceIdString := md[traceIdHeader][0]
			traceId, err := trace.TraceIDFromHex(traceIdString)
			if err != nil {
				log.Error().Err(err).Msg("invalid trace id")
			} else {
				spanContext := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: traceId,
				})
				ctx = trace.ContextWithSpanContext(ctx, spanContext)
			}
		}

		ctx, span := tracer.FromCtx(ctx).Start(ctx, info.FullMethod)
		defer span.End()

		traceId := fmt.Sprintf("%s", span.SpanContext().TraceID())
		ctx = metadata.AppendToOutgoingContext(ctx, traceIdHeader, traceId)

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
		}
		return resp, err
	}
}
