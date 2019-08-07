package did_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
)

const testDID = `
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
			"type": "VerifiableCredentialService",
			"serviceEndpoint": "https://example.com/vc/"
		},
		{
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

const testDIDNoAuthentication = `
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
	err := json.Unmarshal([]byte(testDID), &doc)
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
	err := json.Unmarshal([]byte(testDID), &doc)
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
		if key.Type != "RsaVerificationKey2018" {
			t.Errorf("Should have returned correct type")
		}
	}

	if doc.Authentication == nil {
		t.Errorf("Should have returned authentication items")
	} else {
		if len(doc.Authentication) == 0 {
			t.Errorf("Should have returned some auth items")
		}
		// This is the string id pointer
		key := doc.Authentication[0]
		if key.ID.String() != "did:example:123456789abcdefghi#keys-1" {
			t.Errorf("Should have returned id pointer")
		}
		// This is the full public key object
		key = doc.Authentication[1]
		if key.ID.String() != "did:example:123456789abcdefghi#keys-2" {
			t.Errorf("Should have returned auth key id")
		}
		if key.Type != "RsaVerificationKey2018" {
			t.Errorf("Should have returned correct type for auth key id")
		}
	}
	if doc.Service == nil {
		t.Errorf("Should have returned service items")
	} else {
		if len(doc.Service) == 0 {
			t.Errorf("Should have returned some auth items")
		}
		// URI string
		key := doc.Service[0]
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
		key = doc.Service[1]
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
	err := json.Unmarshal([]byte(testDIDNoAuthentication), &doc)
	if err != nil {
		t.Errorf("Should have unmarshalled document from json: err: %v", err)
	}

	if doc.Authentication != nil {
		t.Errorf("Should have not gotten authentication items")
	}

	if len(doc.Authentication) > 0 {
		t.Errorf("Should have not gotten authentication items")
	}
}

func TestDocumentModelStringify(t *testing.T) {
	doc := did.Document{}
	err := json.Unmarshal([]byte(testDID), &doc)
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
