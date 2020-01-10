// +build liveresolver

package did_test

// Enable the running universal resolver either locally via Docker or via
// SSH tunnel.  Set up to port 8888 on localhost.
// kubectl port-forward deployment/unir-uni-resolver-web 8888:8080 --namespace=staging

import (
	"encoding/json"
	"testing"

	cnum "github.com/joincivil/go-common/pkg/numbers"
	cstr "github.com/joincivil/go-common/pkg/strings"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

const (
	resolverHost = "localhost"
	resolverPort = 8888
)

func TestHTTPUniversalResolverLiveBadDid(t *testing.T) {
	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

	dd, _ := didlib.Parse("did:web:idontexist.co")
	doc, err := res.Resolve(dd)
	if err == nil {
		t.Fatalf("Should have gotten error resolving did")
	}
	if errors.Cause(err) != did.ErrResolverDIDNotFound {
		t.Fatalf("Should have gotten resolver did not found err: %v", errors.Cause(err))
	}
	if doc != nil {
		t.Fatalf("Should have returned empty doc")
	}
}

func TestHTTPUniversalResolverLiveWeb(t *testing.T) {
	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

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

	bys, _ := json.MarshalIndent(doc, "", "    ")
	t.Logf("ethr = %v", string(bys))
}

func TestHTTPUniversalResolverLiveEthr(t *testing.T) {
	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

	dd, _ := didlib.Parse("did:ethr:0x3b0BC51Ab9De1e5B7B6E34E5b960285805C41736")
	doc, err := res.Resolve(dd)
	if err != nil {
		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have received a valid doc")
	}

	bys, _ := json.MarshalIndent(doc, "", "    ")
	t.Logf("ethr = %v", string(bys))
}

// func TestHTTPUniversalResolverLiveCcp(t *testing.T) {
// 	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

// 	dd, _ := didlib.Parse("did:ccp:ceNobbK6Me9F5zwyE3MKY88QZLw")
// 	doc, err := res.Resolve(dd)
// 	if err != nil {
// 		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
// 	}
// 	if doc == nil {
// 		t.Errorf("Should have received a valid doc")
// 	}

// 	bys, _ := json.MarshalIndent(doc, "", "    ")
// 	t.Logf("ccp = %v", string(bys))
// }

// NOTE(PN): DID library is based on old ABNF spec for DIDs, so the underscore in
// nacl fails via this lib.  Going to ping the library maintainers about it or potentially
// fork it for Civil.

// func TestHTTPUniversalResolverLiveNacl(t *testing.T) {
// 	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

// 	dd, err := didlib.Parse("did:nacl:Md8JiMIwsapml_FtQ2ngnGftNP5UmVCAUuhnLyAsPxI")
// 	t.Logf("err = %v", err)
// 	doc, err := res.Resolve(dd)
// 	if err != nil {
// 		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
// 	}
// 	if doc == nil {
// 		t.Errorf("Should have received a valid doc")
// 	}

// 	bys, _ := json.MarshalIndent(doc, "", "    ")
// 	t.Logf("nacl = %v", string(bys))
// }

// NOTE(PN): The resolver for this method doesn't seem to work
// func TestHTTPUniversalResolverLiveSov(t *testing.T) {
// 	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

// 	dd, _ := didlib.Parse("did:sov:WRfXPg8dantKVubE3HX8pw")
// 	doc, err := res.Resolve(dd)
// 	if err != nil {
// 		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
// 	}
// 	if doc == nil {
// 		t.Errorf("Should have received a valid doc")
// 	}

// 	bys, _ := json.MarshalIndent(doc, "", "    ")
// 	t.Logf("sov = %v", string(bys))
// }

// NOTE(PN): The resolver for this method doesn't seem to work
// func TestHTTPUniversalResolverLiveBtcr(t *testing.T) {
// 	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

// 	dd, _ := didlib.Parse("did:btcr:xz35-jznz-q6mr-7q6")
// 	doc, err := res.Resolve(dd)
// 	if err != nil {
// 		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
// 	}
// 	if doc == nil {
// 		t.Errorf("Should have received a valid doc")
// 	}

// 	bys, _ := json.MarshalIndent(doc, "", "    ")
// 	t.Logf("ethr = %v", string(bys))
// }

// NOTE(PN): The resolver for this method doesn't seem to work
// func TestHTTPUniversalResolverLiveWork(t *testing.T) {
// 	res := did.NewHTTPUniversalResolver(cstr.StrToPtr(resolverHost), cnum.IntToPtr(resolverPort), nil)

// 	dd, _ := didlib.Parse("did:work:2UUHQCd4psvkPLZGnWY33L")
// 	doc, err := res.Resolve(dd)
// 	if err != nil {
// 		t.Fatalf("Should not have gotten error resolving did: err: %v", err)
// 	}
// 	if doc == nil {
// 		t.Errorf("Should have received a valid doc")
// 	}

// 	bys, _ := json.MarshalIndent(doc, "", "    ")
// 	t.Logf("ethr = %v", string(bys))
// }
