package chiwares

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

// BasicAuth verifies provided in header username and password
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			u, p, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(401)
				return
			}
			if u != username {
				w.WriteHeader(401)
				return
			}
			if p != password {
				w.WriteHeader(401)
				return
			}

			next.ServeHTTP(ww, r)
		})
	}
}
