package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

const (
	didHeader       = "x-idhub-did"
	reqTsHeader     = "x-idhub-reqts"
	signatureHeader = "x-idhub-signature"
)

var (
	// DidCtxKey is the context key for the auth did
	DidCtxKey = &contextKey{"didkey"}
	// ReqTsCtxKey is the context key for the time stamp
	ReqTsCtxKey = &contextKey{"reqts"}
	// SignatureCtxKey is the key for the signature
	SignatureCtxKey = &contextKey{"signature"}
	// GracePeriodCtxKey is the key for the grace period
	GracePeriodCtxKey = &contextKey{"graceperiod"}
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
			ctx := context.WithValue(r.Context(), DidCtxKey, did)
			ctx = context.WithValue(ctx, ReqTsCtxKey, reqTs)
			ctx = context.WithValue(ctx, SignatureCtxKey, signature)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)

		})
	}
}

// ForContextData is returned by ForContext and contains data pulled from the
// context
type ForContextData struct {
	Did string
}

// ForContext checks signature based on the header data. If error returned,
// indicates an invalid signature or no auth passed. Returns auth context data
// for convenience.
// REQUIRES Middleware to have run.
func ForContext(ctx context.Context, ds *did.Service, pks []did.DocPublicKey) (
	*ForContextData, error) {
	// NOTE(PN): Supporting only Secp251k1 keys for authentication for now
	keyType := linkeddata.SuiteTypeSecp256k1Verification

	reqTs, _ := ctx.Value(ReqTsCtxKey).(string)
	if reqTs == "" {
		return nil, errors.New("no request ts passed in context")
	}
	signature, _ := ctx.Value(SignatureCtxKey).(string)
	if signature == "" {
		return nil, errors.New("no signature passed in context")
	}
	didStr, _ := ctx.Value(DidCtxKey).(string)

	ts, err := strconv.Atoi(reqTs)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert ts to int")
	}

	// altered grace period value can be passed in the context
	gracePeriod := DefaultRequestGracePeriodSecs
	gp, ok := ctx.Value(GracePeriodCtxKey).(int)
	if ok {
		gracePeriod = gp
	}
	// If did and key found, then pull doc for DID to check the signature
	// If no did and key passed, then check incoming list of pks to check signature
	if didStr != "" {
		err = VerifyEcdsaRequestSignatureWithDid(ds, keyType, signature, ts, didStr, gracePeriod)
		if err != nil {
			return nil, err
		}

	} else if pks != nil {
		err = VerifyEcdsaRequestSignatureWithPks(pks, keyType, signature, ts, "", gracePeriod)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, errors.New("could not verify signature, no did or public keys")
	}

	return &ForContextData{Did: didStr}, nil
}
