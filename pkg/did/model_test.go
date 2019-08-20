package did_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
	didlib "github.com/ockam-network/did"
)

const testDIDDoc = `
{
	"@context": "https://w3id.org/did/v1",
	"id": "did:example:123456789abcdefghi",
	"controller": "did:example:123456789abcdefghi",
	"publicKey": [
		{
		"id": "did:example:123456789abcdefghi#keys-1",
		"type": "RsaVerificationKey2018",
		"controller": "did:example:123456789abcdefghi",
		"publicKeyPem": "-----BEGIN PUBLIC KEY...END PUBLIC KEY-----\r\n"
		}
	],
	"authentication": [
		"did:example:123456789abcdefghi#keys-1",
		{
			"id": "did:example:123456789abcdefghi#keys-2",
			"type": "RsaVerificationKey2018",
			"controller": "did:example:123456789abcdefghi",
			"publicKeyPem": "-----BEGIN PUBLIC KEY...END PUBLIC KEY-----\r\n"
		}
	],
	"service": [
		{
	  		"id": "did:example:123456789abcdefghi#vcr",
			"type": "VerifiableCredentialService",
			"serviceEndpoint": "https://example.com/vc/"
		},
		{
	  		"id": "did:example:123456789abcdefghi#hub",
			"type": "IdentityHub",
			"serviceEndpoint": {
				"@context": "https://schema.identity.foundation/hub",
				"type": "UserHubEndpoint",
				"instances": ["did:example:456", "did:example:789"]
			}
		}
	],
	"proof": {
		"type": "LinkedDataSignature2015",
		"created": "2016-02-08T16:02:20Z",
		"creator": "did:example:8uQhQMGzWxR8vw5P3UWH1ja#keys-1",
		"signatureValue": "QNB13Y7Q9...1tzjn4w=="
	},
	"created": "2016-10-17T02:41:00Z",
	"updated": "2016-10-17T02:41:00Z"
}
`

const testDIDDocNoAuthentication = `
{
	"@context": "https://w3id.org/did/v1",
	"id": "did:example:123456789abcdefghi",
	"controller": "did:example:123456789abcdefghi",
	"publicKey": [
		{
		"id": "did:example:123456789abcdefghi#keys-1",
		"type": "RsaVerificationKey2018",
		"controller": "did:example:123456789abcdefghi",
		"publicKeyPem": "-----BEGIN PUBLIC KEY...END PUBLIC KEY-----\r\n"
		}
	],
	"service": [{
	  "id": "did:example:123456789abcdefghi#vcr",
	  "type": "VerifiableCredentialService",
	  "serviceEndpoint": "https://example.com/vc/"
	}],
	"proof": {
	  "type": "LinkedDataSignature2015",
	  "created": "2016-02-08T16:02:20Z",
	  "creator": "did:example:8uQhQMGzWxR8vw5P3UWH1ja#keys-1",
	  "signatureValue": "QNB13Y7Q9...1tzjn4w=="
	},
	"created": "2016-10-17T02:41:00Z",
	"updated": "2016-10-17T02:41:00Z"
}
`

func TestDocumentModelMarshal(t *testing.T) {
	doc := did.Document{}
	err := json.Unmarshal([]byte(testDIDDoc), &doc)
	if err != nil {
		t.Errorf("Should have unmarshalled document from json: err: %v", err)
	}

	bys, err := json.Marshal(&doc)
	if err != nil {
		t.Errorf("Should have marshalled document: err: %v", err)
	}

	jsonStr := string(bys)

	if !strings.Contains(jsonStr, "@context") {
		t.Errorf("Should have contained @context")
	}
	if !strings.Contains(jsonStr, "authentication") {
		t.Errorf("Should have contained authentication")
	}
	if !strings.Contains(jsonStr, "publicKey") {
		t.Errorf("Should have contained publicKey")
	}
	if !strings.Contains(jsonStr, "created") {
		t.Errorf("Should have contained created")
	}
	if !strings.Contains(jsonStr, "updated") {
		t.Errorf("Should have contained updated")
	}
	if !strings.Contains(jsonStr, "controller") {
		t.Errorf("Should have contained controller")
	}

	t.Logf("jsonstr = %v", jsonStr)
}

func TestDocumentModelUnmarshal(t *testing.T) {
	doc := &did.Document{}
	err := json.Unmarshal([]byte(testDIDDoc), &doc)
	if err != nil {
		t.Errorf("Should have unmarshalled document from json: err: %v", err)
	}

	if doc.ID.String() != "did:example:123456789abcdefghi" {
		t.Errorf("Should have returned the correct ID")
	}

	if doc.Controller == nil {
		t.Errorf("Should have had a non-nil controller")
	}

	if doc.Controller != nil && doc.Controller.String() != "did:example:123456789abcdefghi" {
		t.Errorf("Should have returned the correct controller")
	}

	if doc.Context != "https://w3id.org/did/v1" {
		t.Errorf("Should have returned the correct context")
	}

	if doc.PublicKeys == nil {
		t.Errorf("Should have returned public keys")
	} else {
		if len(doc.PublicKeys) == 0 {
			t.Errorf("Should have returned some public keys")
		}
		key := doc.PublicKeys[0]
		if key.ID.String() != "did:example:123456789abcdefghi#keys-1" {
			t.Errorf("Should have returned correct key id")
		}
		if key.Controller.String() != "did:example:123456789abcdefghi" {
			t.Errorf("Should have returned correct controller")
		}
		if key.Type != "RsaVerificationKey2018" {
			t.Errorf("Should have returned correct type")
		}
	}

	if doc.Authentications == nil {
		t.Errorf("Should have returned authentication items")
	} else {
		if len(doc.Authentications) == 0 {
			t.Errorf("Should have returned some auth items")
		}
		// This is the string id pointer
		key := doc.Authentications[0]
		if key.ID.String() != "did:example:123456789abcdefghi#keys-1" {
			t.Errorf("Should have returned id pointer")
		}
		// This is the full public key object
		key = doc.Authentications[1]
		if key.ID.String() != "did:example:123456789abcdefghi#keys-2" {
			t.Errorf("Should have returned auth key id")
		}
		if key.Type != "RsaVerificationKey2018" {
			t.Errorf("Should have returned correct type for auth key id")
		}
	}
	if doc.Services == nil {
		t.Errorf("Should have returned service items")
	} else {
		if len(doc.Services) == 0 {
			t.Errorf("Should have returned some auth items")
		}
		// URI string
		key := doc.Services[0]
		if key.ServiceEndpoint.(string) != "https://example.com/vc/" {
			t.Errorf("Should have valid URI for service endpoint")
		}
		if key.ServiceEndpointURI == nil {
			t.Errorf("Should have set the service endpoint URI")
		}
		if *key.ServiceEndpointURI != "https://example.com/vc/" {
			t.Errorf("Should have valid URI for service endpoint URI")
		}
		// Map for JSONLD
		key = doc.Services[1]
		ld := key.ServiceEndpoint.(map[string]interface{})
		if ld["@context"] != "https://schema.identity.foundation/hub" {
			t.Errorf("Should have valid context for service endpoint ld")
		}
		if ld["type"] != "UserHubEndpoint" {
			t.Errorf("Should have valid type for service endpoint ld")
		}
		if key.ServiceEndpointLD == nil {
			t.Errorf("Should have set the service endpoint LD")
		}
		if key.ServiceEndpointLD["type"] != "UserHubEndpoint" {
			t.Errorf("Should have valid type for service endpoint LD")
		}
	}

	if doc.Created.Year() != 2016 {
		t.Errorf("Should have returned created year")
	}
	if doc.Created.Day() != 17 {
		t.Errorf("Should have returned created day")
	}
	if doc.Created.Month() != 10 {
		t.Errorf("Should have returned created month")
	}
}

func TestDocumentModelUnmarshalNoAuth(t *testing.T) {
	doc := &did.Document{}
	err := json.Unmarshal([]byte(testDIDDocNoAuthentication), &doc)
	if err != nil {
		t.Errorf("Should have unmarshalled document from json: err: %v", err)
	}

	if doc.Authentications != nil {
		t.Errorf("Should have not gotten authentication items")
	}

	if len(doc.Authentications) > 0 {
		t.Errorf("Should have not gotten authentication items")
	}
}

func TestDocumentModelStringify(t *testing.T) {
	doc := did.Document{}
	err := json.Unmarshal([]byte(testDIDDoc), &doc)
	if err != nil {
		t.Errorf("Should have unmarshalled document from json: err: %v", err)
	}

	stringified := fmt.Sprintf("%v", doc)
	if !strings.Contains(stringified, "Document:") {
		t.Errorf("Should have returned the correct string")
	}
	if !strings.Contains(stringified, "created") {
		t.Errorf("Should have returned the correct string")
	}
	if !strings.Contains(stringified, "updated") {
		t.Errorf("Should have returned the correct string")
	}
	t.Logf("string = %v", stringified)
}

func TestNextKeyFragment(t *testing.T) {
	doc := did.Document{}
	err := json.Unmarshal([]byte(testDIDDoc), &doc)
	if err != nil {
		t.Errorf("Should have unmarshalled document from json: err: %v", err)
	}

	frag := doc.NextKeyFragment()
	if frag != "keys-3" {
		t.Errorf("Should have gotten keys-3, got %v", frag)
	}

	// Empty keys, should get keys-1
	doc = did.Document{}
	frag = doc.NextKeyFragment()
	if frag != "keys-1" {
		t.Errorf("Should have gotten keys-1, got %v", frag)
	}
}

func TestAddPublicKey(t *testing.T) {
	doc := did.Document{}
	d, _ := didlib.Parse("did:example:123456789abcdefghi")
	doc.ID = *d

	firstPK := &did.DocPublicKey{
		ID:              doc.ID,
		Type:            did.LDSuiteTypeSecp256k1Verification,
		Controller:      &doc.ID,
		EthereumAddress: "0x5E4A048a9B8F5256a0D485e86E31e2c3F86523FB",
	}

	// Adding first PK and authentication
	err := doc.AddPublicKey(firstPK.SetIDFragment(doc.NextKeyFragment()), true)
	if err != nil {
		t.Errorf("Should have added first public key")
	}

	if len(doc.PublicKeys) != 1 {
		t.Errorf("Should have 1 key")
	}

	secondPK := &did.DocPublicKey{
		ID:              doc.ID,
		Type:            did.LDSuiteTypeSecp256k1Verification,
		Controller:      &doc.ID,
		EthereumAddress: "0xf5a27f027125f07fef36871db3c0f68015370589",
	}

	// Adding second PK
	err = doc.AddPublicKey(secondPK.SetIDFragment(doc.NextKeyFragment()), false)
	if err != nil {
		t.Errorf("Should have added second public key")
	}

	if len(doc.PublicKeys) != 2 {
		t.Errorf("Should have 2 keys")
	}
	pk2 := doc.PublicKeys[1]
	if !strings.HasSuffix(pk2.ID.String(), "#keys-2") {
		t.Errorf("Should have keys-2 fragment")
	}

	thirdPK := &did.DocPublicKey{
		ID:              doc.ID,
		Type:            did.LDSuiteTypeSecp256k1Verification,
		Controller:      &doc.ID,
		EthereumAddress: "0xdad6d7ea1e43f8492a78bab8bb0d45a889ed6ac3",
	}

	// Adding third PK and second authentication
	err = doc.AddPublicKey(thirdPK.SetIDFragment(doc.NextKeyFragment()), true)
	if err != nil {
		t.Errorf("Should have added second public key")
	}

	if len(doc.PublicKeys) != 3 {
		t.Errorf("Should have 3 keys")
	}
	pk2 = doc.PublicKeys[2]
	t.Logf("pk = %v\n", pk2.ID.String())
	if !strings.HasSuffix(pk2.ID.String(), "#keys-3") {
		t.Errorf("Should have keys-3 fragment")
	}
	if len(doc.Authentications) != 2 {
		t.Errorf("Should have 2 authentications")
	}
	auth2 := doc.Authentications[1]
	if !strings.HasSuffix(auth2.ID.String(), "#keys-3") {
		t.Errorf("Auth should have keys-3 fragment")
	}

	d, _ = didlib.Parse("did:example:testme#keys-1")
	fourthPK := &did.DocPublicKey{
		ID:              *d,
		Type:            did.LDSuiteTypeSecp256k1Verification,
		Controller:      &doc.ID,
		EthereumAddress: "0xdad6d7ea1e43f8492a78bab8bb0d45a889ed6ac3",
	}

	// Adding fourth PK
	err = doc.AddPublicKey(fourthPK.SetIDFragment(doc.NextKeyFragment()), false)
	if err != nil {
		t.Errorf("Should have added second public key")
	}
	if len(doc.PublicKeys) != 4 {
		t.Errorf("Should have 4 keys")
	}
	pk2 = doc.PublicKeys[3]
	if !strings.HasSuffix(pk2.ID.String(), "#keys-4") {
		t.Errorf("Should have keys-4 fragment")
	}

	bys, _ := json.Marshal(doc)
	t.Logf("%v", string(bys))
}
