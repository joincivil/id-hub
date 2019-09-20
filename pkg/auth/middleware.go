package auth

import (
	"net/http"
)

// Middleware for auth handles authorization of
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO(PN): Do some shit here
			next.ServeHTTP(w, r)
		})
	}
}
