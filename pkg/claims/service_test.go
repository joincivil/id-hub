package claims_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

func setupConnection() (*gorm.DB, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}
	db.DropTable(&did.PostgresDocument{})
	err = db.AutoMigrate(&did.PostgresDocument{}, &claimsstore.SignedClaimPostgres{}, &claimsstore.Node{}).Error
	if err != nil {
		return nil, err
	}
	return db, nil
}

func makeContentCredential(issuerDID *didlib.DID) *claimsstore.ContentCredential {
	subj := claimsstore.ContentCredentialSubject{
		ID: "https://ap.com/article/1",
		Metadata: article.Metadata{
			Title: "something something",
		},
	}
	proof := linkeddata.Proof{}
	return &claimsstore.ContentCredential{
		Context:           []string{"https://something.com/some/stuff/v1"},
		Type:              []claimsstore.CredentialType{claimsstore.VerifiableCredentialType, claimsstore.ContentCredentialType},
		CredentialSubject: subj,
		Issuer:            issuerDID.String(),
		IssuanceDate:      time.Date(2018, 2, 1, 12, 30, 0, 0, time.UTC),
		Proof:             proof,
	}
}

func addProof(cred *claimsstore.ContentCredential, signerDID *didlib.DID, pk *ecdsa.PrivateKey) error {
	canonical, err := claims.CanonicalizeCredential(cred)
	if err != nil {
		return err
	}
	hash := crypto.Keccak256(canonical)
	sigBytes, err := crypto.Sign(hash, pk)
	if err != nil {
		return err
	}

	proofValue := hex.EncodeToString(sigBytes)
	cred.Proof = linkeddata.Proof{
		Type:       string(linkeddata.SuiteTypeSecp256k1Signature),
		Creator:    signerDID.String(),
		Created:    time.Now(),
		ProofValue: proofValue,
	}
	return nil
}

func makeService(db *gorm.DB, didService *did.Service, signedClaimStore *claimsstore.SignedClaimPGPersister) (*claims.Service, error) {
	nodepersister := claimsstore.NewNodePGPersisterWithDB(db)
	treeStore := claimsstore.NewPGStore(nodepersister)
	claimService, err := claims.NewService(treeStore, signedClaimStore, didService)
	return claimService, err
}

func TestCreateTreeForDID(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(userDIDs)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}
	secKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey := secKey.Public().(*ecdsa.PublicKey)
	err = claimService.CreateTreeForDID(userDID, []*ecdsa.PublicKey{pubKey})
	if err != nil {
		t.Errorf("error creating tree for did: %v", err)
	}
	didClaims, err := claimService.GetMerkleTreeClaimsForDid(userDID)
	if err != nil {
		t.Errorf("error getting claims for user: %v", err)
	}
	if len(didClaims) != 1 {
		t.Errorf("there should be one claim in the users tree")
	}
	rootClaims, err := claimService.GetRootMerkleTreeClaims()
	if err != nil {
		t.Errorf("error getting root claims: %v", err)
	}
	if len(rootClaims) != 1 {
		t.Errorf("there should be one claim in the root tree")
	}
}

func TestTreeExistsForDID(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	testDID := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(testDID)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}
	secKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey := secKey.Public().(*ecdsa.PublicKey)

	// Tree doesn't exist yet, should return false
	exists, err := claimService.TreeExistsForDID(userDID)
	if err != nil {
		t.Errorf("should not have returned error: err: %v", err)
	}
	if exists {
		t.Errorf("should not have returned exists = true")
	}

	// Create the tree.
	err = claimService.CreateTreeForDID(userDID, []*ecdsa.PublicKey{pubKey})
	if err != nil {
		t.Errorf("error creating tree for did: %v", err)
	}

	// Tree exists now, should return true
	exists, err = claimService.TreeExistsForDID(userDID)
	if err != nil {
		t.Errorf("should not have returned error: err: %v", err)
	}
	if !exists {
		t.Errorf("should not have returned exists = false")
	}
}

func TestCreateTreeForDIDIfNotExists(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	testDID := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(testDID)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}

	// Return error since the user DID doesn't exist
	err = claimService.CreateTreeForDIDIfNotExists(userDID)
	if err == nil {
		t.Errorf("should have returned an error since did does not exist")
	}

	// Create a test DID document
	key, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Fatalf("should be able to make a key")
	}
	pubBytes := crypto.FromECDSAPub(&key.PublicKey)
	pub := hex.EncodeToString(pubBytes)
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	docPubKey.ID = did.CopyDID(userDID)
	docPubKey.Controller = did.CopyDID(userDID)
	didDoc, err := did.InitializeNewDocument(userDID, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := didService.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	// Return no error since the user DID exists
	err = claimService.CreateTreeForDIDIfNotExists(userDID)
	if err != nil {
		t.Errorf("should not have returned error: err: %v", err)
	}
}

func TestClaimContent(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}
	key, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Fatalf("should be able to make a key")
	}
	pubBytes := crypto.FromECDSAPub(&key.PublicKey)
	pub := hex.EncodeToString(pubBytes)
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

	cred := makeContentCredential(&didDoc.ID)
	err = addProof(cred, didDoc.PublicKeys[0].ID, key)
	if err != nil {
		t.Errorf("error adding proof: %v", err)
	}

	err = claimService.ClaimContent(cred)
	if err == nil {
		t.Errorf("should have errored because couldn't resolv the key")
	}
	err = claimService.CreateTreeForDID(&didDoc.ID, []*ecdsa.PublicKey{&key.PublicKey})
	if err != nil {
		t.Errorf("problem creating did tree: %v", err)
	}
	err = claimService.ClaimContent(cred)
	if err != nil {
		t.Errorf("problem creating content claim: %v", err)
	}
	err = claimService.ClaimContent(cred)
	if err == nil {
		t.Errorf("should err for duplicate claim")
	}
	cred.Proof.ProofValue = "04e9627daa1419d73a7a3bdd9e907a9bf0ae4344149521d4b5d07377b589658265e705971b26da6d51bbea4ef7ecf5267f10437126add370f752a1b2f0af65c32f"
	err = claimService.ClaimContent(cred)
	if err == nil {
		t.Errorf("should have errored for the bad signature")
	}
	listDidClaims, err := claimService.GetMerkleTreeClaimsForDid(&didDoc.ID)
	if err != nil {
		t.Errorf("error retrieving claims from did tree: %v", err)
	}
	listRootClaims, err := claimService.GetRootMerkleTreeClaims()
	if err != nil {
		t.Errorf("error retrieving claims from root tree: %v", err)
	}
	if len(listDidClaims) != 2 {
		t.Errorf("unexpected number of did claims")
	}
	if len(listRootClaims) != 2 {
		t.Errorf("unexpected number of did claims")
	}

	for _, v := range listDidClaims {
		switch v.(type) {
		case claims.ClaimRegisteredDocument:
			regClaim := v.(claims.ClaimRegisteredDocument)
			claimHash := hex.EncodeToString(regClaim.ContentHash[:])
			signedClaim, err := signedClaimStore.GetCredentialByHash(claimHash)
			if err != nil {
				t.Errorf("could not retrieve credential: %v", err)
			}
			if signedClaim.CredentialSubject.ID != "https://ap.com/article/1" {
				t.Errorf("unexpected value for credential")
			}
		}
	}
}

func TestClaimsToContentCredentials(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()

	// Setup
	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	// Create a DID identity
	key, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Fatalf("should be able to make a key")
	}
	pubBytes := crypto.FromECDSAPub(&key.PublicKey)
	pub := hex.EncodeToString(pubBytes)
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

	// Create the DID tree
	ecdsaPubkey, _ := crypto.UnmarshalPubkey(pubBytes)
	err = claimService.CreateTreeForDID(&didDoc.ID, []*ecdsa.PublicKey{ecdsaPubkey})
	if err != nil {
		t.Errorf("problem creating did tree: %v", err)
	}

	// Claim content 1
	cred := makeContentCredential(&didDoc.ID)
	_ = addProof(cred, didDoc.PublicKeys[0].ID, key)
	err = claimService.ClaimContent(cred)
	if err != nil {
		t.Errorf("problem creating content claim: %v", err)
	}

	listDidClaims, err := claimService.GetMerkleTreeClaimsForDid(&didDoc.ID)
	if err != nil {
		t.Errorf("error retrieving claims from did tree: %v", err)
	}

	contentCreds, err := claimService.ClaimsToContentCredentials(listDidClaims)
	if err != nil {
		t.Errorf("error converting claims to content creds: %v", err)
	}

	if len(listDidClaims) == 2 && len(contentCreds) != 1 {
		t.Errorf("should have filtered down to 1 content cred from 2 claims")
	}
}
