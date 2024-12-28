package chiwares

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader Overriding WriteHeader to capture the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code) // Forward to the original ResponseWriter
}

// Default status code is 200 if WriteHeader isn't called
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

type PrometheusMiddleware struct {
	excludeRoutes map[string]struct{}

	httpRequestDuration *prometheus.HistogramVec
	httpRequestsTotal   *prometheus.CounterVec
}

// NewPrometheusMiddleware returns a new instance of Prometheus middleware.
// excludeRoutes is a list of routes that should be excluded from metrics. By default, /metrics is excluded.
func NewPrometheusMiddleware(excludeRoutes []string) *PrometheusMiddleware {
	// Convert excludeRoutes to a map for efficient lookups
	excludeMap := make(map[string]struct{}, len(excludeRoutes))
	for _, route := range excludeRoutes {
		excludeMap[route] = struct{}{}
	}

	// Exclude /metrics by default
	excludeMap["/metrics"] = struct{}{}

	// Define Prometheus metrics
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response time for HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Register metrics
	prometheus.MustRegister(httpRequestDuration, httpRequestsTotal)

	return &PrometheusMiddleware{
		excludeRoutes:       excludeMap,
		httpRequestDuration: httpRequestDuration,
		httpRequestsTotal:   httpRequestsTotal,
	}
}

// Handle returns http.Handler that writes response metrics to Prometheus
func (m *PrometheusMiddleware) Handle() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := m.excludeRoutes[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			status := rw.statusCode

			// Record metrics
			m.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, http.StatusText(status)).Observe(duration)
			m.httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(status)).Inc()
		})
	}
}
