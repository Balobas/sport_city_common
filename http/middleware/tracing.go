package httpMiddleware

import (
	"fmt"
	"net/http"

	"github.com/balobas/sport_city_common/tracer"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/metadata"
)

func Tracing() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rctx := chi.RouteContext(r.Context())

			ctx, span := tracer.FromCtx(ctx).Start(ctx, rctx.RoutePattern())
			span.SetAttributes(
				attribute.String("url", r.URL.String()),
				attribute.String("method", r.Method),
				attribute.String("reqId", middleware.GetReqID(r.Context())),
			)
			defer span.End()

			traceId := fmt.Sprintf("%s", span.SpanContext().TraceID())
			ctx = metadata.AppendToOutgoingContext(ctx, traceIdHeader, traceId)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

const traceIdHeader = "x-trace-id"
