package chiwares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	t.Run("should log request and response", func(t *testing.T) {
		// Given
		var logOutput bytes.Buffer
		logger := zerolog.New(&logOutput)

		r := chi.NewRouter()
		r.Use(Logger(logger))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// When
		r.ServeHTTP(rr, req)

		// Then
		assert.Equal(t, http.StatusOK, rr.Code)
		logContent := logOutput.String()
		assert.Contains(t, logContent, "request started")
		assert.Contains(t, logContent, "request finished")
		assert.Contains(t, logContent, `"method":"GET"`)
		assert.Contains(t, logContent, `"path":"/"`)
		assert.Contains(t, logContent, `"status":200`)
	})

	t.Run("should ignore specified paths", func(t *testing.T) {
		// Given
		var logOutput bytes.Buffer
		logger := zerolog.New(&logOutput)

		r := chi.NewRouter()
		r.Use(Logger(logger, WithIgnorePaths([]string{"/health"})))
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// When
		reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
		rrHealth := httptest.NewRecorder()
		r.ServeHTTP(rrHealth, reqHealth)

		// Then
		assert.Equal(t, http.StatusOK, rrHealth.Code)
		assert.Equal(t, "", logOutput.String())

		// When
		reqHome := httptest.NewRequest(http.MethodGet, "/", nil)
		rrHome := httptest.NewRecorder()
		r.ServeHTTP(rrHome, reqHome)

		// Then
		assert.Equal(t, http.StatusOK, rrHome.Code)
		logContent := logOutput.String()
		assert.Contains(t, logContent, "request started")
		assert.Contains(t, logContent, "request finished")
	})

	t.Run("should log slow requests", func(t *testing.T) {
		// Given
		var logOutput bytes.Buffer
		logger := zerolog.New(&logOutput)
		durationThreshold := 50 * time.Millisecond

		r := chi.NewRouter()
		r.Use(Logger(logger, WithDurationThreshold(durationThreshold)))
		r.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(durationThreshold + 10*time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		rr := httptest.NewRecorder()

		// When
		r.ServeHTTP(rr, req)

		// Then
		assert.Equal(t, http.StatusOK, rr.Code)
		logContent := logOutput.String()
		assert.True(t, strings.Contains(logContent, "slow request"), "Expected log to contain 'slow request'")
	})
}
