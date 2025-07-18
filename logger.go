package chiwares

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5/middleware"
)

// LoggerOption is a function that configures the logger middleware.
type LoggerOption func(*loggerOptions)

// loggerOptions holds the configuration for the logger middleware.
type loggerOptions struct {
	durationThreshold time.Duration
	ignorePaths       []string
}

// WithDurationThreshold sets the duration threshold for slow requests.
func WithDurationThreshold(threshold time.Duration) LoggerOption {
	return func(o *loggerOptions) {
		o.durationThreshold = threshold
	}
}

// WithIgnorePaths sets the paths to be ignored by the logger.
func WithIgnorePaths(paths []string) LoggerOption {
	return func(o *loggerOptions) {
		o.ignorePaths = paths
	}
}

// Logger is a middleware that logs request and response
// append logger middleware after RequestID middleware
func Logger(logger zerolog.Logger, opts ...LoggerOption) func(http.Handler) http.Handler {
	options := &loggerOptions{}
	for _, o := range opts {
		o(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range options.ignorePaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			startTS := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = fmt.Sprintf("%d", rand.Int())
			}

			l := logger.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).Logger()

			defer func() {
				duration := time.Since(startTS)
				lEntry := l.With().
					Str("duration", duration.String()).
					Int("status", ww.Status()).Logger()

				lEntry.Debug().Msg("request finished")

				if options.durationThreshold > 0 && duration > options.durationThreshold {
					lEntry.Warn().Any("http_request", r).Msg("slow request")
				}
			}()

			l.Debug().Msg("request started")
			next.ServeHTTP(ww, r)
		})
	}
}
