package auth

import (
	"context"
	"net/http"

	log "github.com/golang/glog"

	"github.com/joincivil/id-hub/pkg/did"
)

const (
	didKeyHeader    = "x-idhub-did-key"
	reqTsHeader     = "x-idhub-req-ts"
	signatureHeader = "x-idhub-signature"
)

var (
	didKeyCtxKey    = &contextKey{"didKey"}
	reqTsCtxKey     = &contextKey{"reqTs"}
	signatureCtxKey = &contextKey{"signature"}
)

type contextKey struct {
	name string
}

// Middleware for auth handles authorization based on public key and/or DID.
func Middleware(didService *did.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Put incoming request auth headers into the context or use later
			// to verify signature
			reqTs := r.Header.Get(reqTsHeader)
			signature := r.Header.Get(signatureHeader)
			didKey := r.Header.Get(didKeyHeader)

			// didKey is optional
			if reqTs == "" && signature == "" {
				log.Infof("No auth headers found")
				next.ServeHTTP(w, r)
				return
			}

			// Store values into context
			ctx := context.WithValue(r.Context(), didKeyCtxKey, didKey)
			ctx = context.WithValue(ctx, reqTsCtxKey, reqTs)
			ctx = context.WithValue(ctx, signatureCtxKey, signature)

			// and call the next with our new context
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
