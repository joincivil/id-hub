package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/did"
)

const (
	didHeader       = "x-idhub-did"
	reqTsHeader     = "x-idhub-reqts"
	signatureHeader = "x-idhub-signature"
)

var (
	didCtxKey       = &contextKey{"didkey"}
	reqTsCtxKey     = &contextKey{"reqts"}
	signatureCtxKey = &contextKey{"signature"}
)

type contextKey struct {
	name string
}

// Middleware for auth handles authorization based on public key and/or DID.
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Put incoming request auth headers into the context or use later
			// to verify signature
			reqTs := r.Header.Get(reqTsHeader)
			signature := r.Header.Get(signatureHeader)
			did := r.Header.Get(didHeader)

			// didKey is optional
			if reqTs == "" && signature == "" {
				log.Infof("No auth headers found")
				next.ServeHTTP(w, r)
				return
			}

			// Store values into context
			ctx := context.WithValue(r.Context(), didCtxKey, did)
			ctx = context.WithValue(ctx, reqTsCtxKey, reqTs)
			ctx = context.WithValue(ctx, signatureCtxKey, signature)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

		})
	}
}

// ForContext checks signature based on the header data.
// REQUIRES Middleware to have run.
func ForContext(ctx context.Context, ds *did.Service, pks []did.DocPublicKey) error {
	reqTs, _ := ctx.Value(reqTsCtxKey).(string)
	signature, _ := ctx.Value(signatureCtxKey).(string)

	// NOTE(PN): Supporting only Secp251k1 keys for authentication for now
	keyType := did.LDSuiteTypeSecp256k1Verification
	didStr, _ := ctx.Value(didCtxKey).(string)

	ts, err := strconv.Atoi(reqTs)
	if err != nil {
		return errors.Wrap(err, "could not convert ts to int")
	}

	// If did and key found, then pull doc for DID to check the signature
	// If no did and key passed, then check incoming list of pks to check signature
	if didStr != "" {
		err = VerifyEcdsaRequestSignatureWithDid(ds, keyType, signature, ts, didStr)
		if err != nil {
			return err
		}

	} else if pks != nil {
		err = VerifyEcdsaRequestSignatureWithPks(pks, keyType, signature, ts, "")
		if err != nil {
			return err
		}

	} else {
		return errors.New("could not verify signature, no did or public keys")
	}

	return nil
}
