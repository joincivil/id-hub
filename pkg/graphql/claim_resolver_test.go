package graphql_test

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/go-common/pkg/lock"
	ctime "github.com/joincivil/go-common/pkg/time"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

var claimJSON = `{
	"@context":["https://something.com/some/stuff/v1"],
	"type":["VerifiableCredential","ContentCredential"],
	"credentialSubject":{
		"id":"https://ap.com/article/1",
		"metadata":{
			"Title":"something something",
			"RevisionContentHash":"",
			"RevisionContentURL":"",
			"CanonicalURL":"",
			"Slug":"",
			"Description":"",
			"Contributors":null,
			"Images":null,
			"Tags":null,
			"PrimaryTag":"",
			"RevisionDate":"0001-01-01T00:00:00Z",
			"OriginalPublishDate":"0001-01-01T00:00:00Z",
			"Opinion":false,
			"CivilSchemaVersion":""
		}
	},
	"issuer":"did:ethuri:cc4ef0ec-bd37-46e6-8419-3164c325205f",
	"credentialSchema":{
		"id":"",
		"type":""
	},
	"issuanceDate":"2018-02-01T12:30:00Z",
	"proof":[{
		"type":"EcdsaSecp256k1Signature2019",
		"creator":"did:ethuri:cc4ef0ec-bd37-46e6-8419-3164c325205f#keys-1",
		"created":"2019-10-09T17:09:39.753902-05:00",
		"proofValue":"a91c9cb42f277475696bee83090b8ec72f2903e908e5c0058689039a29cbafdd65babf18898e5272f180a203c6817ec72a934e68356d1d9b7a783127ca7465b101"
	}]
}`

func initPersister(t *testing.T) (ethuri.Persister, *gorm.DB) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		t.Fatalf("Should have returned a new gorm db conn")
		return nil, nil
	}

	err = db.AutoMigrate(&ethuri.PostgresDocument{}).Error
	if err != nil {
		t.Fatalf("Should have auto-migrated")
		return nil, nil
	}

	return ethuri.NewPostgresPersister(db), db
}

func initEthURIService(t *testing.T) (*did.Service, *ethuri.Service, *gorm.DB) {
	persister, db := initPersister(t)
	ethURIService := ethuri.NewService(persister)
	didService := did.NewService([]did.Resolver{ethURIService})
	return didService, ethURIService, db
}

func makeService(db *gorm.DB, didService *did.Service, signedClaimStore *claimsstore.SignedClaimPGPersister) (
	*claims.Service, *claims.RootService, *claims.JWTService, error) {
	nodepersister := claimsstore.NewNodePGPersisterWithDB(db)
	treeStore := claimsstore.NewPGStore(nodepersister)
	rootCommitStore := claimsstore.NewRootCommitsPGPersister(db)
	dlock := lock.NewLocalDLock()

	committer := &claims.FakeRootCommitter{CurrentBlockNumber: big.NewInt(2)}
	rootService, _ := claims.NewRootService(treeStore, committer, rootCommitStore)

	claimService, err := claims.NewService(treeStore, signedClaimStore, didService, rootService, dlock)
	didJWTService := didjwt.NewService(didService)
	jwtStore := claimsstore.NewJWTClaimPGPersister(db, didJWTService)
	jwtService := claims.NewJWTService(didJWTService, jwtStore, claimService, &testutils.FakeNatsService{})
	return claimService, rootService, jwtService, err
}

func createDID(service *ethuri.Service, claimerDid *didlib.DID) error {
	pubKeyHex := "046d94c84a7096c572b83d44df576e1ffb3573123f62099f8d4fa19de806bd4d5939d36f91cc5e69398b5709f184abae4c128664b024bddfd09585de74bd85cdbf"
	pubk := &did.DocPublicKey{
		ID:           claimerDid,
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pubKeyHex,
		Controller:   claimerDid,
	}
	doc, err := ethuri.InitializeNewDocument(claimerDid, pubk, true, true)
	if err != nil {
		return err
	}
	return service.SaveDocument(doc)
}

func setupResolver(t *testing.T) (*graphql.Resolver, *claims.RootService, *ethuri.Service, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, nil, nil, err
	}
	db.DropTable(&ethuri.PostgresDocument{}, &claimsstore.RootCommit{}, &claimsstore.Node{}, &claimsstore.SignedClaimPostgres{})
	err = db.AutoMigrate(&ethuri.PostgresDocument{}, &claimsstore.SignedClaimPostgres{}, &claimsstore.Node{}, &claimsstore.RootCommit{}).Error
	if err != nil {
		return nil, nil, nil, err
	}
	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didService, ethURI, db := initEthURIService(t)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, rootService, jwtService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		return nil, nil, nil, err
	}

	resolver := &graphql.Resolver{
		DidService:   didService,
		ClaimService: claimService,
		JWTService:   jwtService,
	}

	return resolver, rootService, ethURI, nil
}

func TestClaimSaveAndProof(t *testing.T) {
	resolver, rootService, ethURI, err := setupResolver(t)

	if err != nil {
		t.Errorf("couldn't set up resolver")
	}

	claimerDid, err := didlib.Parse("did:ethuri:cc4ef0ec-bd37-46e6-8419-3164c325205f")
	if err != nil {
		t.Errorf("couldn't parse did: %v", err)
	}

	if err := createDID(ethURI, claimerDid); err != nil {
		t.Errorf("couldn't add the did: %v", err)
	}

	cred := &claimtypes.ContentCredential{}
	err = json.Unmarshal([]byte(claimJSON), cred)
	if err != nil {
		t.Errorf("couldn't make cred: %v", err)
	}

	pk, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Errorf("couldn't create private key: %v", err)
	}

	claimSaveInput := &graphql.ClaimSaveRequestInput{
		ClaimJSON: &claimJSON,
	}

	claimerDid.Fragment = ""

	reqTs := ctime.CurrentEpochSecsInInt()

	signature, err := auth.SignEcdsaRequestMessage(pk, claimerDid.String(), reqTs)
	if err != nil {
		t.Errorf("couldn't create auth signature: %v", err)
	}

	queries := resolver.Query()
	mutations := resolver.Mutation()

	_, err = mutations.ClaimSave(context.Background(), claimSaveInput)
	if err == nil {
		t.Errorf("should have errored on auth")
	}

	c := context.Background()
	c = context.WithValue(c, auth.ReqTsCtxKey, strconv.Itoa(reqTs))
	c = context.WithValue(c, auth.DidCtxKey, claimerDid.String())
	c = context.WithValue(c, auth.SignatureCtxKey, signature)

	claimSaveRes, err := mutations.ClaimSave(c, claimSaveInput)
	if err != nil {
		t.Errorf("unexpected err save the claim: %v", err)
	}

	if claimSaveRes.Claim.Issuer != cred.Issuer {
		t.Errorf("unexpected return from claimsave")
	}

	// commit the root
	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing root: %v", err)
	}

	proofInput := &graphql.ClaimProofRequestInput{
		Claim:     claimSaveInput.Claim,
		ClaimJSON: claimSaveInput.ClaimJSON,
		Did:       claimerDid.String(),
	}

	proofResponse, err := queries.ClaimProof(c, proofInput)
	if err != nil {
		t.Errorf("error generating proof: %v", err)
	}

	if len(proofResponse.Claim.Proof.([]interface{})) != 3 {
		t.Errorf("wrong number of proofs on claim")
	}
}
