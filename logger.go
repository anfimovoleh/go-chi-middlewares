package chiwares

import (
	"fmt"
	"github.com/rs/zerolog"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// TODO: add HTTP tests

// Logger is a middleware that logs request and response
// append logger middleware after RequestID middleware
func Logger(logger zerolog.Logger, durationThreshold time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTS := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = fmt.Sprintf("%d", rand.Int())
			}

			logger = logger.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).Logger()

			defer func() {
				duration := time.Since(startTS)
				lEntry := logger.With().
					Str("duration", duration.String()).
					Int("status", ww.Status()).Logger()

				lEntry.Debug().Msg("request finished")

				if duration > durationThreshold {
					lEntry.Warn().Any("http_request", r).Msg("slow request")
				}
			}()

			logger.Debug().Msg("request started")
			next.ServeHTTP(ww, r)
		})
	}
}
