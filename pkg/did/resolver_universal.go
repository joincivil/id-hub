package did

// A client for the universal DID resolver defined and implemented by the DIF
// https://github.com/decentralized-identity/universal-resolver/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/ockam-network/did"

	chttp "github.com/joincivil/go-common/pkg/http"
)

const (
	defaultResolverHost = "uni-resolver-web"
	defaultResolverPort = 8080
	uniResolverURL      = "http://%v:%v/1.0/identifiers/%v"

	reqMaxAtts    = 3
	reqBaseWaitMs = 50
)

// UniversalResolverResponse is the response from the universal resolver
type UniversalResolverResponse struct {
	DidDocument    *Document                  `json:"didDocument"`
	Metadata       *UniversalResolverMetadata `json:"resolverMetadata"`
	Content        *string                    `json:"content"`
	ContentType    *string                    `json:"contentType"`
	MethodMetadata *map[string]interface{}    `json:"methodMetadata"`
}

// UniversalResolverMetadata is the metadata response for the universal resolver response
type UniversalResolverMetadata struct {
	Duration   int    `json:"duration"`
	Identifier string `json:"identifier"`
	DriverID   string `json:"driverId"`
	DidURL     struct {
		DidURLString string `json:"didUrlString"`
		Did          struct {
			DidString        string `json:"didString"`
			Method           string `json:"method"`
			MethodSpecificID string `json:"methodSpecificId"`
			ParseTree        string `json:"parseTree"`
			ParseRuleCount   string `json:"parseRuleCount"`
		} `json:"did"`
		Parameters     string                 `json:"parameters"`
		ParametersMap  map[string]interface{} `json:"parametersMap"`
		Path           string                 `json:"path"`
		Query          string                 `json:"query"`
		Fragment       string                 `json:"fragment"`
		ParseTree      string                 `json:"parseTree"`
		ParseRuleCount string                 `json:"parseRuleCount"`
	} `json:"didUrl"`
}

// NewHTTPUniversalResolver initializes and returns a new HTTPUniversalResolver
func NewHTTPUniversalResolver(resolverHost *string, resolverPort *int,
	cache ResolverCache) *HTTPUniversalResolver {
	host := defaultResolverHost
	if resolverHost != nil {
		host = *resolverHost
	}
	port := defaultResolverPort
	if resolverPort != nil {
		port = *resolverPort
	}
	return &HTTPUniversalResolver{
		resolverHost: host,
		resolverPort: port,
		cache:        cache,
	}
}

// HTTPUniversalResolver implements Resolver for making requests to the
// identityfoundation/universal-resolver service.
type HTTPUniversalResolver struct {
	resolverHost string
	resolverPort int
	cache        ResolverCache
}

// Resolve returns the DID document given the DID
// Implements the Resolver interface.
func (h *HTTPUniversalResolver) Resolve(d *did.DID) (*Document, error) {
	if h.cache != nil {
		doc, err := h.cache.Get(d)
		if err == nil && doc != nil {
			return doc, nil
		}

		if err != nil && errors.Cause(err) != ErrResolverCacheDIDNotFound {
			return nil, errors.Wrap(err, "resolve.get")
		}
	}

	resp, err := h.RawResolve(d)
	if err != nil {
		return nil, errors.Wrap(err, "resolve.rawresolve")
	}

	if h.cache != nil && resp != nil {
		err = h.cache.Set(d, resp.DidDocument)
		if err != nil {
			return nil, errors.Wrap(err, "resolve.set")
		}
	}

	return resp.DidDocument, nil
}

// RawResolve returns the full universal resolver resp given the DID
func (h *HTTPUniversalResolver) RawResolve(d *did.DID) (*UniversalResolverResponse, error) {
	if d == nil {
		return nil, errors.New("Invalid DID")
	}

	client := chttp.NewRestHelper("", "")
	res, err := client.SendRequestToURLWithRetry(
		h.fullResolverURL(d),
		http.MethodGet,
		nil,
		nil,
		reqMaxAtts,
		reqBaseWaitMs,
	)
	if err != nil {
		cause := errors.Cause(err)
		if strings.Contains(strings.ToLower(cause.Error()),
			"resolve problem for") {
			log.Infof("Resolver err: %v", cause.Error())
			return nil, ErrResolverDIDNotFound
		}

		return nil, errors.Wrap(err, "resolve.sendrequest")
	}

	resp := &UniversalResolverResponse{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, errors.Wrap(err, "resolve.unmarshal")
	}

	return resp, nil
}

func (h *HTTPUniversalResolver) fullResolverURL(d *did.DID) string {
	return fmt.Sprintf(uniResolverURL, h.resolverHost, h.resolverPort, d.String())
}
