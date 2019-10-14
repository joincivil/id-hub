// +build creds

package claims_test

// This are to manually test saving claims to an ID Hub with authentication.
// NOTE: An ID Hub should be running on localhost:8080 and postgresql should be up.
// go test -cpu 4 -v -race -timeout=180s -tags=creds -logtostderr=true -stderrthreshold=INFO -run TestCredentialsAndGqlClaimSave

// Manually ensure the DB is cleaned up since the server and this test mutates data.

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	gql "github.com/machinebox/graphql"

	ctime "github.com/joincivil/go-common/pkg/time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/linkeddata"

	didlib "github.com/ockam-network/did"
)

const (
	testDid = "did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1"
)

func testClaimSaveGQL(t *testing.T, credStr string, reqTs int, thedid string,
	sig string) error {
	client := gql.NewClient("http://localhost:8080/v1/query")
	req := gql.NewRequest(`
		mutation($in: ClaimSaveRequestInput!) {
			claimSave(in: $in) {
				claimRaw
			}
		}
	`)

	req.Header.Add("X-IDHUB-DID", testDid)
	req.Header.Add("X-IDHUB-REQTS", strconv.Itoa(reqTs))
	req.Header.Add("X-IDHUB-SIGNATURE", sig)

	req.Var("in", graphql.ClaimSaveRequestInput{ClaimJSON: &credStr})

	resp := struct {
		ClaimRaw string `json:"claimRaw"`
	}{}
	err := client.Run(context.Background(), req, &resp)
	if err != nil {
		t.Logf("err running graphql")
		return err
	}

	t.Logf("returned from gql: %+v", resp.ClaimRaw)
	return nil
}

func testServerUp() error {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		return err
	} else {
		conn.Close()
	}
	return nil
}

// Used to generated some predetermined test creds
func TestCredentialsAndGqlClaimSave(t *testing.T) {
	err := testServerUp()
	if err != nil {
		t.Fatalf("server not up, please start up the ID Hub at 8080, err: %v", err)
	}

	// Setup DB
	db, err := setupConnection()
	if err != nil {
		t.Fatalf("error setting up the db: %v", err)
	}

	// Setup services
	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Fatalf("error setting up service: %v", err)
	}

	// New private key
	privKey, _ := crypto.GenerateKey()
	privBys := crypto.FromECDSA(privKey)
	priv := hex.EncodeToString(privBys)

	// New pub key
	pubKey := privKey.PublicKey
	pubBys := crypto.FromECDSAPub(&pubKey)
	pub := hex.EncodeToString(pubBys)

	t.Logf("priv key hex:\n%v", priv)
	t.Logf("pub key hex:\n%v", pub)
	t.Logf("did:\n%v", testDid)

	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	// Create DID and DID document
	signerDid, err := didlib.Parse(testDid)
	if err != nil {
		t.Fatalf("error creating did: %v", err)
	}
	docPubKey.ID = did.CopyDID(signerDid)
	docPubKey.Controller = did.CopyDID(signerDid)
	didDoc, err := did.InitializeNewDocument(signerDid, docPubKey, true, true)
	if err != nil {
		t.Fatalf("error making the did doc: %v", err)
	}
	err = didService.SaveDocument(didDoc)
	if err != nil {
		t.Fatalf("error saving the did doc: %v", err)
	}
	t.Logf("saved did document")

	// New claims tree for the DID
	err = claimService.CreateTreeForDID(&didDoc.ID, &pubKey)
	if err != nil {
		t.Fatalf("problem creating did tree: %v", err)
	}

	bys, _ := json.MarshalIndent(didDoc, "", "    ")
	t.Logf("test did:\n%v", signerDid)
	t.Logf("test did doc:\n%v", string(bys))

	// Create a new content credential / claim
	cred := makeContentCredential(&didDoc.ID)
	canoncred, _ := claims.CanonicalizeCredential(cred)
	fmt.Printf("canoncred = %v\n", string(canoncred))

	// Create proof on credential
	proofValue, _ := auth.SignMessage(privKey, canoncred)
	fmt.Printf("proof val = %v\n", proofValue)
	cred.Proof = linkeddata.Proof{
		Type:       string(linkeddata.SuiteTypeSecp256k1Signature),
		Creator:    didDoc.PublicKeys[0].ID.String(),
		Created:    time.Now(),
		ProofValue: proofValue,
	}

	// Sign the request message using the private key
	reqTs := ctime.CurrentEpochSecsInInt()
	sig, err := auth.SignEcdsaRequestMessage(privKey, testDid, reqTs)
	if err != nil {
		t.Fatalf("unable to sign message: err: %v", err)
	}

	bys, _ = json.Marshal(cred)
	t.Logf("req sig:\n%v\n", sig)
	t.Logf("test cred:\n%v\n", string(bys))

	// time.Sleep(time.Millisecond * 2000)

	// Send the GQL request
	t.Logf("calling gql")
	err = testClaimSaveGQL(t, string(bys), reqTs, testDid, sig)
	if err != nil {
		t.Fatalf("error calling gql: %v", err)
	}
}
