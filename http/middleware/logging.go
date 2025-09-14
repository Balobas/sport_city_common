package httpMiddleware

import (
	"net/http"
	"time"

	"github.com/balobas/sport_city_common/logger"
	"github.com/go-chi/chi/middleware"
)

func Logging() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			log := logger.Logger()

			reqId := middleware.GetReqID(r.Context())

			log = log.With().Str("req_id", reqId).Logger()
			r = r.WithContext(logger.ContextWithLogger(r.Context(), log))

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			log.Info().
				Timestamp().
				Fields(map[string]interface{}{
					"remote_ip":  r.RemoteAddr,
					"url":        r.URL.Path,
					"proto":      r.Proto,
					"method":     r.Method,
					"user_agent": r.Header.Get("User-Agent"),
				}).Msg("incoming request")
			t1 := time.Now()

			next.ServeHTTP(ww, r)

			t2 := time.Now()

			// log end request
			log.Info().
				Timestamp().
				Fields(map[string]interface{}{
					"url":        r.URL.Path,
					"method":     r.Method,
					"status":     ww.Status(),
					"latency_ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
				}).
				Msg("incoming request finished")
		}
		return http.HandlerFunc(fn)
	}
}
