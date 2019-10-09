// +build creds

package claims_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	didlib "github.com/ockam-network/did"
)

// Used to generated some predetermined test creds
func TestMakeTestCredentials(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	pub := "049691d8097f07afb7068a971ba500abd30b2ef763240bc56bf021ff592ed08446b7d23df1a5a043e7472d8954764f3fd39fbf992517e9c61ba10afee1965391e6"
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	signerDid, err := didlib.Parse("did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1")
	if err != nil {
		t.Errorf("error creating did: %v", err)
	}
	docPubKey.ID = signerDid
	docPubKey.Controller = did.CopyDID(signerDid)
	didDoc, err := did.InitializeNewDocument(signerDid, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := didService.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	pubBytes, _ := hex.DecodeString(pub)
	ecdsaPubkey, _ := crypto.UnmarshalPubkey(pubBytes)
	err = claimService.CreateTreeForDID(&didDoc.ID, ecdsaPubkey)
	if err != nil {
		t.Errorf("problem creating did tree: %v", err)
	}

	bys, _ := json.MarshalIndent(didDoc, "", "    ")

	t.Logf("test did:\n%v", signerDid)
	t.Logf("test doc:\n%v", string(bys))

	cred := makeContentCredential(&didDoc.ID)
	addProof(cred, didDoc.PublicKeys[0].ID)

	bys, _ = json.MarshalIndent(cred, "", "    ")
	t.Logf("test cred:\n%v\n", string(bys))
}
