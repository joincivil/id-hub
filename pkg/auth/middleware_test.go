package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
)

func initService(t *testing.T) *did.Service {
	persister := &did.InMemoryPersister{}
	return did.NewService(persister)
}

type testHandler struct {
	t            *testing.T
	checkHeaders bool
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))

	if h.checkHeaders {
		didKey := r.Header.Get(didKeyHeader)
		if didKey == "" {
			h.t.Errorf("Should have set didKey in header")
		}
		reqTs := r.Header.Get(reqTsHeader)
		if reqTs == "" {
			h.t.Errorf("Should have set reqTs in header")
		}
		signature := r.Header.Get(signatureHeader)
		if signature == "" {
			h.t.Errorf("Should have set signature in header")
		}
	}
}

func TestMiddleware(t *testing.T) {
	service := initService(t)

	middleware := Middleware(service)
	server1 := httptest.NewServer(middleware(&testHandler{t: t, checkHeaders: false}))

	// No headers
	_, _ = http.Get(server1.URL)

	server2 := httptest.NewServer(middleware(&testHandler{t: t, checkHeaders: true}))

	// With the headers
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server2.URL, nil)
	req.Header.Add(didKeyHeader, "did:ethurl:123456")
	req.Header.Add(reqTsHeader, "12344567")
	req.Header.Add(signatureHeader, "thisisasignature")

	_, _ = client.Do(req)
}
