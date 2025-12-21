package middleware

import (
	"net/http"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
)

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac := auth.FromRequest(r)
		r = r.WithContext(auth.WithContext(r.Context(), ac))
		next.ServeHTTP(w, r)
	})
}
