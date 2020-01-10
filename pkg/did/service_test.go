package did_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	cnum "github.com/joincivil/go-common/pkg/numbers"
	cstr "github.com/joincivil/go-common/pkg/strings"
	"github.com/joincivil/id-hub/pkg/did"
	didlib "github.com/ockam-network/did"
)

type NoResolutionResolver struct {
}

func (n *NoResolutionResolver) Resolve(d *didlib.DID) (*did.Document, error) {
	return nil, did.ErrResolverDIDNotFound
}

func TestGetDocument(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	serv := did.NewService([]did.Resolver{res})

	doc, err := serv.GetDocument("did:web:uport.me")
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}

	if doc.ID.String() != "did:web:uport.me" {
		t.Errorf("Should have gotten the correct ID")
	}

	doc, err = serv.GetDocument("did")
	if err == nil {
		t.Errorf("Should have gotten error")
	}
	if doc != nil {
		t.Errorf("Should have gotten empty doc")
	}
}

func TestGetDocumentNoResolution(t *testing.T) {
	res := &NoResolutionResolver{}

	serv := did.NewService([]did.Resolver{res})

	doc, err := serv.GetDocument("did:web:idontexist.co")
	if err == nil {
		t.Errorf("Should have gotten error")
	}
	if doc != nil {
		t.Errorf("Should have gotten empty doc")
	}
	if err != did.ErrResolverDIDNotFound {
		t.Errorf("Should have gotten did not found error")
	}
}

func TestGetDocumentFromDID(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	serv := did.NewService([]did.Resolver{res})

	dd, _ := didlib.Parse("did:web:uport.me")

	doc, err := serv.GetDocumentFromDID(dd)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}

	if doc.ID.String() != "did:web:uport.me" {
		t.Errorf("Should have gotten the correct ID")
	}
}

func TestGetKeyFromDIDDocument(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	serv := did.NewService([]did.Resolver{res})

	// Test normal scenario
	dd, _ := didlib.Parse("did:web:uport.me#owner")
	key, err := serv.GetKeyFromDIDDocument(dd)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}

	if key.ID.String() != "did:web:uport.me#owner" {
		t.Errorf("Should have gotten the correct ID")
	}

	// Test no fragment
	dd, _ = didlib.Parse("did:web:uport.me")
	_, err = serv.GetKeyFromDIDDocument(dd)
	if err == nil {
		t.Errorf("Should have gotten error: err: %v", err)
	}
}

func TestGetKeyFromDIDDocumentNoDID(t *testing.T) {
	res := &NoResolutionResolver{}
	serv := did.NewService([]did.Resolver{res})

	// Test no did
	dd, _ := didlib.Parse("did:web:civil.co#owner")
	_, err := serv.GetKeyFromDIDDocument(dd)
	if err == nil {
		t.Errorf("Should have gotten error: err: %v", err)
	}
}
