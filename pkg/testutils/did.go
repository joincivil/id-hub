package testutils

import (
	"fmt"

	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/utils"
)

const (
	testDID = "did:ethuri:fbaf6bb3-2a82-4173-b31a-160a143c931c"
)

// BuildTestDocument builds a test DID document
func BuildTestDocument() *did.Document {
	doc := &did.Document{}

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
	hexKey := "04f3df3cea421eac2a7f5dbd8e8d505470d42150334f512bd6383c7dc91bf8fa4d5458d498b4dcd05574c902fb4c233005b3f5f3ff3904b41be186ddbda600580b"
	pk1.PublicKeyHex = utils.StrToPtr(hexKey)

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
