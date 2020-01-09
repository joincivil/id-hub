package did_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/allegro/bigcache"
	cnum "github.com/joincivil/go-common/pkg/numbers"
	cstr "github.com/joincivil/go-common/pkg/strings"

	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

const (
	validResponse = `{
		"didDocument": {
			"@context": "https://w3id.org/did/v1",
			"id": "did:web:uport.me",
			"service": [],
			"authentication": [
				{
					"type": "Secp256k1SignatureAuthentication2018",
					"publicKey": [
						"did:web:uport.me#owner"
					]
				}
			],
			"publicKey": [
				{
					"id": "did:web:uport.me#owner",
					"type": "Secp256k1VerificationKey2018",
					"owner": "did:web:uport.me",
					"publicKeyHex": "042b0af9b3ae6c7c3a90b01a3879d9518081bc0dcdf038488db9cb109b082a77d97ea3373e3dfde0eccd9adbdce11d0302ea5c098dbb0b310234c86895c8641622"
				}
			]
		},
		"resolverMetadata": {
			"duration": 1349,
			"identifier": "did:web:uport.me",
			"driverId": "driver-uport/uni-resolver-driver-did-uport",
			"didUrl": {
				"didUrlString": "did:web:uport.me",
				"did": {
					"didString": "did:web:uport.me",
					"method": "web",
					"methodSpecificId": "uport.me",
					"parseTree": null,
					"parseRuleCount": null
				},
				"parameters": null,
				"parametersMap": {},
				"path": "",
				"query": null,
				"fragment": null,
				"parseTree": null,
				"parseRuleCount": null
			}
		},
		"methodMetadata": {},
		"content": null,
		"contentType": null
	}`
	invalidResponse = `{
		"didDocument": {
			"@context": "https://w3id.org/did/v1"
		},
		"resolverMetadata": {
			"duration": 1349,
			"identifier": "did:web:uport.me",
			"driverId": "driver-uport/uni-resolver-driver-did-uport",
			"didUrl": {
				"didUrlString": "did:web:uport.me",
				"did": {
					"didString": "did:web:uport.me",
					"method": "web",
					"methodSpecificId": "uport.me",
					"parseTree": null,
					"parseRuleCount": null
				},
				"parameters": null,
				"parametersMap": {},
				"path": "",
				"query": null,
				"fragment": null,
				"parseTree": null,
				"parseRuleCount": null
			}
		},
		"methodMetadata": {},
		"content": null,
		"contentType": null
	}`
)

func TestHTTPUniversalResolver(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	dd, _ := didlib.Parse("did:web:uport.me")
	doc, err := res.Resolve(dd)
	if err != nil {
		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have received a valid doc")
	}

	if doc.ID.String() != "did:web:uport.me" {
		t.Errorf("DID is incorrect")
	}

	if len(doc.Authentications) != 1 {
		t.Errorf("Should have gotten 1 authentication entry")
	}

	if doc.Authentications[0].Type != linkeddata.SuiteTypeSecp256k1SignatureAuth2018 {
		t.Errorf("Should have gotten 1 authentication entry")
	}

	if len(doc.Authentications[0].PublicKey) != 1 {
		t.Fatalf("Should have gotten 1 publickey in auth entry")
	}

	if doc.Authentications[0].PublicKey[0] != "did:web:uport.me#owner" {
		t.Fatalf("Should have gotten correct auth public key")
	}

	if len(doc.PublicKeys) != 1 {
		t.Fatalf("Should have gotten 1 public key")
	}

	if doc.PublicKeys[0].ID.String() != "did:web:uport.me#owner" {
		t.Fatalf("Should have gotten correct public key")
	}

	if doc.PublicKeys[0].Type != linkeddata.SuiteTypeSecp256k1Verification2018 {
		t.Fatalf("Should have gotten correct public key type")
	}

	if doc.PublicKeys[0].Owner.String() != "did:web:uport.me" {
		t.Fatalf("Should have gotten correct public key owner")
	}

	server.Close()
}

func TestHTTPUniversalResolverWithCache(t *testing.T) {
	count := 0
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count > 1 {
			t.Error("Should have been cached")
		}
		t.Logf("here")
		fmt.Fprintln(w, validResponse)
		count++
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	cache, _ := bigcache.NewBigCache(bigcache.Config{
		Shards:             1024,
		LifeWindow:         2 * time.Second,
		CleanWindow:        3 * time.Second,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		Verbose:            true,
		HardMaxCacheSize:   16384,
	})
	rcache := did.NewBigCacheResolverCache(cache)

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), rcache)

	dd, _ := didlib.Parse("did:web:uport.me")
	doc, err := res.Resolve(dd)
	if err != nil {
		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have received a valid doc")
	}

	doc, err = res.Resolve(dd)
	if err != nil {
		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have received a valid doc")
	}

	doc, err = res.Resolve(dd)
	if err != nil {
		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have received a valid doc")
	}

	server.Close()
}

func TestHTTPUniversalRawResolver(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	dd, _ := didlib.Parse("did:web:uport.me")
	resp, err := res.RawResolve(dd)

	if resp.Metadata.Duration != 1349 {
		t.Fatalf("Should have gotten 1349")
	}
	if resp.Metadata.DriverID != "driver-uport/uni-resolver-driver-did-uport" {
		t.Fatalf("Should have gotten the right driver ID")
	}

	doc := resp.DidDocument

	if err != nil {
		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have received a valid doc")
	}

	if doc.ID.String() != "did:web:uport.me" {
		t.Errorf("DID is incorrect")
	}

	if len(doc.Authentications) != 1 {
		t.Errorf("Should have gotten 1 authentication entry")
	}

	if doc.Authentications[0].Type != linkeddata.SuiteTypeSecp256k1SignatureAuth2018 {
		t.Errorf("Should have gotten 1 authentication entry")
	}

	if len(doc.Authentications[0].PublicKey) != 1 {
		t.Fatalf("Should have gotten 1 publickey in auth entry")
	}

	if doc.Authentications[0].PublicKey[0] != "did:web:uport.me#owner" {
		t.Fatalf("Should have gotten correct auth public key")
	}

	if len(doc.PublicKeys) != 1 {
		t.Fatalf("Should have gotten 1 public key")
	}

	if doc.PublicKeys[0].ID.String() != "did:web:uport.me#owner" {
		t.Fatalf("Should have gotten correct public key")
	}

	if doc.PublicKeys[0].Type != linkeddata.SuiteTypeSecp256k1Verification2018 {
		t.Fatalf("Should have gotten correct public key type")
	}

	if doc.PublicKeys[0].Owner.String() != "did:web:uport.me" {
		t.Fatalf("Should have gotten correct public key owner")
	}

	server.Close()
}

func TestHTTPUniversalResolverErrorEmptyDID(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	resp, err := res.Resolve(nil)
	if err == nil {
		t.Errorf("Should have returned an error")
	}
	if resp != nil {
		t.Errorf("Should have returned an empty response")
	}
}

func TestHTTPUniversalResolverErrorHttp(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	dd, _ := didlib.Parse("did:web:uport.me")
	resp, err := res.Resolve(dd)
	if err == nil {
		t.Errorf("Should have returned an error")
	}
	if resp != nil {
		t.Errorf("Should have returned an empty response")
	}
}

func TestHTTPUniversalRawResolverError(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, invalidResponse)
	})

	server := httptest.NewServer(h)
	u, _ := url.Parse(server.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(host), cnum.IntToPtr(port), nil)

	dd, _ := didlib.Parse("did:web:uport.me")
	resp, err := res.Resolve(dd)
	if err == nil {
		t.Errorf("Should have returned an error")
	}
	if resp != nil {
		t.Errorf("Should have returned an empty response")
	}
}
