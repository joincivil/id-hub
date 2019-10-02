package auth

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/utils"

	ceth "github.com/joincivil/go-common/pkg/eth"
	ctime "github.com/joincivil/go-common/pkg/time"

	didlib "github.com/ockam-network/did"
)

const (
	testDID = "did:ethuri:fbaf6bb3-2a82-4173-b31a-160a143c931c"
)

func initService() *did.Service {
	persister := &did.InMemoryPersister{}
	return did.NewService(persister)
}

type testHandler struct {
	t            *testing.T
	checkHeaders bool
	checkContext bool
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))

	if h.checkHeaders {
		didKey := r.Header.Get(didHeader)
		if didKey == "" {
			h.t.Errorf("Should have set didKey in header")
		}
		if didKey != "did:ethurl:123456" {
			h.t.Errorf("Should have had matched did key")
		}
		reqTs := r.Header.Get(reqTsHeader)
		if reqTs == "" {
			h.t.Errorf("Should have set reqTs in header")
		}
		if reqTs != "1234567" {
			h.t.Errorf("Should have had matched reqTs value")
		}
		signature := r.Header.Get(signatureHeader)
		if signature == "" {
			h.t.Errorf("Should have set signature in header")
		}
		if signature != "thisisasignature" {
			h.t.Errorf("Should have had matched signature value")
		}
	}

	ctx := r.Context()

	if h.checkContext {
		didKey := ctx.Value(didCtxKey).(string)
		if didKey == "" {
			h.t.Errorf("Should have set didKey in context")
		}
		if didKey != "did:ethurl:123456" {
			h.t.Errorf("Should have had matched did key")
		}
		reqTs := ctx.Value(reqTsCtxKey).(string)
		if reqTs == "" {
			h.t.Errorf("Should have set reqTs in context")
		}
		if reqTs != "1234567" {
			h.t.Errorf("Should have had matched reqTs value")
		}
		signature := ctx.Value(signatureCtxKey).(string)
		if signature == "" {
			h.t.Errorf("Should have set signature in context")
		}
		if signature != "thisisasignature" {
			h.t.Errorf("Should have had matched signature value")
		}
	}
}

func TestMiddleware(t *testing.T) {
	middleware := Middleware()
	server1 := httptest.NewServer(middleware(&testHandler{t: t}))

	// No headers
	_, _ = http.Get(server1.URL)

	server2 := httptest.NewServer(middleware(&testHandler{t: t,
		checkHeaders: true, checkContext: true}))

	// With the headers
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server2.URL, nil)
	req.Header.Add(didHeader, "did:ethurl:123456")
	req.Header.Add(reqTsHeader, "1234567")
	req.Header.Add(signatureHeader, "thisisasignature")

	_, _ = client.Do(req)
}

type testHandlerForContext struct {
	t  *testing.T
	ds *did.Service
}

func (h *testHandlerForContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := ForContext(ctx, h.ds, nil)
	if err != nil {
		h.t.Errorf("Should not have returned a bad verification: err: %v", err)
	}
}

func TestForContext(t *testing.T) {
	middleware := Middleware()
	ds := initService()

	acct, _ := ceth.MakeAccount()
	privKey := acct.Key

	d := buildTestDocument(privKey)

	err := ds.SaveDocument(d)
	if err != nil {
		t.Fatalf("Should have not gotten error saving doc")
	}

	ts := ctime.CurrentEpochSecsInInt()
	server1 := httptest.NewServer(
		middleware(&testHandlerForContext{t: t, ds: ds}),
	)

	signature, err := SignEcdsaRequestMessage(privKey, d.ID.String(), ts)
	if err != nil {
		t.Fatalf("Should have generated a signature")
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server1.URL, nil)
	req.Header.Add(didHeader, d.ID.String())
	req.Header.Add(reqTsHeader, strconv.Itoa(ts))
	req.Header.Add(signatureHeader, signature)

	_, _ = client.Do(req)
}

type testHandlerForContextNewDid struct {
	t   *testing.T
	ds  *did.Service
	pks []did.DocPublicKey
}

func (h *testHandlerForContextNewDid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := ForContext(ctx, h.ds, h.pks)
	if err != nil {
		h.t.Errorf("Should have returned verified: err: %v", err)
	}
}

func TestForContextNewDid(t *testing.T) {
	middleware := Middleware()
	ds := initService()

	acct, _ := ceth.MakeAccount()
	privKey := acct.Key
	pubKeyBys := crypto.FromECDSAPub(&privKey.PublicKey)
	pubKeyHex := hex.EncodeToString(pubKeyBys)

	pks := []did.DocPublicKey{
		{
			Type:         linkeddata.SuiteTypeSecp256k1Verification,
			PublicKeyHex: &pubKeyHex,
		},
	}

	ts := ctime.CurrentEpochSecsInInt()
	server1 := httptest.NewServer(
		middleware(&testHandlerForContextNewDid{
			t:   t,
			ds:  ds,
			pks: pks,
		}),
	)

	signature, err := SignEcdsaRequestMessage(privKey, "", ts)
	if err != nil {
		t.Fatalf("Should have generated a signature: err: %v", err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server1.URL, nil)
	req.Header.Add(reqTsHeader, strconv.Itoa(ts))
	req.Header.Add(signatureHeader, signature)

	_, _ = client.Do(req)
}

type testHandlerForContextNewDidNoPk struct {
	t   *testing.T
	ds  *did.Service
	pks []did.DocPublicKey
}

func (h *testHandlerForContextNewDidNoPk) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := ForContext(ctx, h.ds, h.pks)
	if err == nil {
		h.t.Errorf("Should have not been verified")
	}
}
func TestForContextNewDidNoPk(t *testing.T) {
	middleware := Middleware()
	ds := initService()

	acct, _ := ceth.MakeAccount()
	privKey := acct.Key

	pks := []did.DocPublicKey{}

	ts := ctime.CurrentEpochSecsInInt()
	server1 := httptest.NewServer(
		middleware(&testHandlerForContextNewDidNoPk{
			t:   t,
			ds:  ds,
			pks: pks,
		}),
	)

	signature, err := SignEcdsaRequestMessage(privKey, "", ts)
	if err != nil {
		t.Fatalf("Should have generated a signature: err: %v", err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server1.URL, nil)
	req.Header.Add(reqTsHeader, strconv.Itoa(ts))
	req.Header.Add(signatureHeader, signature)

	_, _ = client.Do(req)
}

type testHandlerForContextBad struct {
	t  *testing.T
	ds *did.Service
}

func (h *testHandlerForContextBad) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := ForContext(ctx, h.ds, nil)
	if err == nil {
		h.t.Errorf("Should have returned a bad verification")
	}
}

func TestForContextBadSignature(t *testing.T) {
	middleware := Middleware()
	ds := initService()

	privKey, _ := crypto.GenerateKey()
	d := buildTestDocument(privKey)

	err := ds.SaveDocument(d)
	if err != nil {
		t.Fatalf("Should have not gotten error saving doc")
	}

	ts := ctime.CurrentEpochSecsInInt()

	privKey2, _ := crypto.GenerateKey()
	signature, err := SignEcdsaRequestMessage(privKey2, d.ID.String(), ts)
	if err != nil {
		t.Fatalf("Should have generated a signature")
	}

	server1 := httptest.NewServer(
		middleware(&testHandlerForContextBad{t: t, ds: ds}),
	)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server1.URL, nil)
	req.Header.Add(didHeader, d.ID.String())
	req.Header.Add(reqTsHeader, strconv.Itoa(ts))
	req.Header.Add(signatureHeader, signature)

	_, _ = client.Do(req)
}

func buildTestDocument(privKey *ecdsa.PrivateKey) *did.Document {
	doc := &did.Document{}
	pubKeyBys := crypto.FromECDSAPub(&privKey.PublicKey)
	pubKeyHex := hex.EncodeToString(pubKeyBys)

	mainDID, _ := didlib.Parse(testDID)

	doc.ID = *mainDID
	doc.Context = did.DefaultDIDContextV1
	doc.Controller = mainDID

	// Public Keys
	pk1 := did.DocPublicKey{}
	pk1ID := fmt.Sprintf("%v#keys-1", testDID)
	d1, _ := didlib.Parse(pk1ID)
	pk1.ID = d1
	pk1.Type = linkeddata.SuiteTypeSecp256k1Verification
	pk1.Controller = mainDID
	pk1.PublicKeyHex = utils.StrToPtr(pubKeyHex)

	doc.PublicKeys = []did.DocPublicKey{pk1}

	// Service endpoints
	ep1 := did.DocService{}
	ep1ID := fmt.Sprintf("%v#vcr", testDID)
	d2, _ := didlib.Parse(ep1ID)
	ep1.ID = *d2
	ep1.Type = "CredentialRepositoryService"
	ep1.ServiceEndpoint = "https://repository.example.com/service/8377464"
	ep1.ServiceEndpointURI = utils.StrToPtr("https://repository.example.com/service/8377464")

	doc.Services = []did.DocService{ep1}

	// Authentication
	aw1 := did.DocAuthenicationWrapper{}
	aw1ID := fmt.Sprintf("%v#keys-1", testDID)
	d3, _ := didlib.Parse(aw1ID)
	aw1.ID = d3
	aw1.IDOnly = true

	aw2 := did.DocAuthenicationWrapper{}
	aw2ID := fmt.Sprintf("%v#keys-2", testDID)
	d4, _ := didlib.Parse(aw2ID)
	aw2.ID = d4
	aw2.IDOnly = false
	aw2.Type = linkeddata.SuiteTypeSecp256k1Verification
	aw2.Controller = mainDID
	hexKey2 := "04debef3fcbef3f5659f9169bad80044b287139a401b5da2979e50b032560ed33927eab43338e9991f31185b3152735e98e0471b76f18897d764b4e4f8a7e8f61b"
	aw2.PublicKeyHex = utils.StrToPtr(hexKey2)

	doc.Authentications = []did.DocAuthenicationWrapper{aw1, aw2}

	return doc
}
