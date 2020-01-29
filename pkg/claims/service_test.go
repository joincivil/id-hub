package claims_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/go-common/pkg/lock"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
	"github.com/multiformats/go-multihash"
	didlib "github.com/ockam-network/did"
)

func initDIDService(db *gorm.DB) (*did.Service, *ethuri.Service) {
	persister := ethuri.NewPostgresPersister(db)
	ethURIService := ethuri.NewService(persister)
	return did.NewService([]did.Resolver{ethURIService}), ethURIService
}

func setupConnection() (*gorm.DB, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}
	db.DropTable(&ethuri.PostgresDocument{}, &claimsstore.RootCommit{}, &claimsstore.Node{})
	err = db.AutoMigrate(&ethuri.PostgresDocument{}, &claimsstore.SignedClaimPostgres{}, &claimsstore.Node{}, &claimsstore.RootCommit{}, &claimsstore.JWTClaimPostgres{}).Error
	if err != nil {
		return nil, err
	}
	return db, nil
}

func makeContentCredential(issuerDID *didlib.DID) *claimtypes.ContentCredential {
	subj := claimtypes.ContentCredentialSubject{
		ID: "https://ap.com/article/1",
		Metadata: article.Metadata{
			Title: "something something",
		},
	}
	proof := linkeddata.Proof{}
	proofSlice := make([]interface{}, 0, 1)
	proofSlice = append(proofSlice, proof)
	return &claimtypes.ContentCredential{
		Context:           []string{"https://something.com/some/stuff/v1"},
		Type:              []claimtypes.CredentialType{claimtypes.VerifiableCredentialType, claimtypes.ContentCredentialType},
		CredentialSubject: subj,
		Issuer:            issuerDID.String(),
		IssuanceDate:      time.Date(2018, 2, 1, 12, 30, 0, 0, time.UTC),
		Proof:             proofSlice,
	}
}

func makeLicenseCredential(issuerDID *didlib.DID, subjectDID *didlib.DID) *claimtypes.LicenseCredential {
	subj1 := &claimtypes.ContentSubject{
		ID:    "https://id.civil.co/contentcredentials/v1/abcde",
		Owner: issuerDID.String(),
	}
	subj2 := &claimtypes.LicenserSubject{
		ID:   subjectDID.String(),
		Name: "Some Publisher",
	}
	credSubj := &[]interface{}{subj1, subj2}
	proof := linkeddata.Proof{}
	proofSlice := &[]interface{}{proof}
	return &claimtypes.LicenseCredential{
		Context:           []string{"https://something.com/some/stuff/v1"},
		Type:              []claimtypes.CredentialType{claimtypes.VerifiableCredentialType, claimtypes.LicenseCredentialType},
		CredentialSubject: credSubj,
		Issuer:            issuerDID.String(),
		IssuanceDate:      time.Date(2018, 2, 1, 12, 30, 0, 0, time.UTC),
		ExpirationDate:    time.Date(2019, 2, 1, 12, 30, 0, 0, time.UTC),
		Proof:             proofSlice,
	}
}

func makeService(db *gorm.DB, didService *did.Service,
	signedClaimStore *claimsstore.SignedClaimPGPersister) (*claims.Service, *claims.RootService, error) {
	nodepersister := claimsstore.NewNodePGPersisterWithDB(db)
	treeStore := claimsstore.NewPGStore(nodepersister)
	rootCommitStore := claimsstore.NewRootCommitsPGPersister(db)
	dlock := lock.NewLocalDLock()
	committer := &claims.FakeRootCommitter{CurrentBlockNumber: big.NewInt(1)}
	rootService, _ := claims.NewRootService(treeStore, committer, rootCommitStore)
	claimService, err := claims.NewService(treeStore, signedClaimStore, didService, rootService, dlock)
	return claimService, rootService, err
}

func TestCreateTreeForDIDWithPks(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didService, _ := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, _, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(userDIDs)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}

	// Create the tree
	secKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey := secKey.Public().(*ecdsa.PublicKey)
	err = claimService.CreateTreeForDIDWithPks(userDID, []*ecdsa.PublicKey{pubKey})
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

	// Try to put the same key in, should fail
	err = claimService.CreateTreeForDIDWithPks(userDID, []*ecdsa.PublicKey{pubKey})
	if err != nil {
		t.Errorf("should not have gotten err for adding an existing key")
	}
	didClaims, err = claimService.GetMerkleTreeClaimsForDid(userDID)
	if err != nil {
		t.Errorf("error getting claims for user: %v", err)
	}
	if len(didClaims) != 1 {
		t.Errorf("there should be one claim in the users tree")
	}

	// Add another key to the tree, make sure using the same tree
	secKey2, err := crypto.HexToECDSA("5dff9479ddd0b9f2213d415b5810fd7e9950ce1312a3c62fa42c6894560197a7")
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey2 := secKey2.Public().(*ecdsa.PublicKey)
	err = claimService.CreateTreeForDIDWithPks(userDID, []*ecdsa.PublicKey{pubKey2})
	if err != nil {
		t.Errorf("error creating tree for did: %v", err)
	}
	didClaims, err = claimService.GetMerkleTreeClaimsForDid(userDID)
	if err != nil {
		t.Errorf("error getting claims for user: %v", err)
	}
	if len(didClaims) != 2 {
		t.Errorf("there should be two claims in the users tree")
	}
	// Should see the total claims for all dids
	rootClaims, err = claimService.GetRootMerkleTreeClaims()
	if err != nil {
		t.Errorf("error getting root claims: %v", err)
	}
	if len(rootClaims) != 2 {
		t.Errorf("there should be two claims in the root tree")
	}
}

func TestCreateTreeForDID(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didService, ethURI := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, _, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	testDID := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(testDID)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}

	// Return error since the user DID doesn't exist
	err = claimService.CreateTreeForDID(userDID)
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
	didDoc, err := ethuri.InitializeNewDocument(userDID, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	// Return no error since the user DID exists
	err = claimService.CreateTreeForDID(userDID)
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
	didService, ethURI := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, _, err := makeService(db, didService, signedClaimStore)
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
	didDoc, err := ethuri.InitializeNewDocument(signerDid, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	cred := makeContentCredential(&didDoc.ID)
	err = claims.AddProof(cred, didDoc.PublicKeys[0].ID, key)
	if err != nil {
		t.Errorf("error adding proof: %v", err)
	}

	err = claimService.ClaimContent(cred)
	if err == nil {
		t.Errorf("should have errored because couldn't resolv the key")
	}
	err = claimService.CreateTreeForDIDWithPks(&didDoc.ID,
		[]*ecdsa.PublicKey{&key.PublicKey})
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
	proofs, ok := cred.Proof.([]interface{})
	if !ok {
		t.Fatalf("proofs is not []interface{}")
	}
	linkedDataProof, ok := proofs[0].(linkeddata.Proof)
	if !ok {
		t.Errorf("should be a linked data proof")
	}

	linkedDataProof.ProofValue = "04e9627daa1419d73a7a3bdd9e907a9bf0ae4344149521d4b5d07377b589658265e705971b26da6d51bbea4ef7ecf5267f10437126add370f752a1b2f0af65c32f"
	proofs[0] = linkedDataProof
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
		switch val := v.(type) {
		case claimtypes.ClaimRegisteredDocument:
			regClaim := val
			claimHash := hex.EncodeToString(regClaim.ContentHash[:])
			signedClaim, err := signedClaimStore.GetCredentialByHash(claimHash)
			if err != nil {
				t.Errorf("could not retrieve credential: %v", err)
			}
			switch signedClaimValue := signedClaim.(type) {
			case *claimtypes.ContentCredential:
				if signedClaimValue.CredentialSubject.ID != "https://ap.com/article/1" {
					t.Errorf("unexpected value for credential")
				}
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
	didService, ethURI := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, _, err := makeService(db, didService, signedClaimStore)
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
	didDoc, err := ethuri.InitializeNewDocument(signerDid, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	// Create the DID tree
	ecdsaPubkey, _ := crypto.UnmarshalPubkey(pubBytes)
	err = claimService.CreateTreeForDIDWithPks(&didDoc.ID,
		[]*ecdsa.PublicKey{ecdsaPubkey})
	if err != nil {
		t.Errorf("problem creating did tree: %v", err)
	}

	// Claim content 1
	cred := makeContentCredential(&didDoc.ID)
	_ = claims.AddProof(cred, didDoc.PublicKeys[0].ID, key)
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

func TestGenerateProof(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()

	// Setup
	didService, ethURI := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, rootService, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	// Create the the did
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
	didDoc, err := ethuri.InitializeNewDocument(signerDid, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	// Create the DID tree
	ecdsaPubkey, _ := crypto.UnmarshalPubkey(pubBytes)
	err = claimService.CreateTreeForDIDWithPks(&didDoc.ID,
		[]*ecdsa.PublicKey{ecdsaPubkey})
	if err != nil {
		t.Errorf("problem creating did tree: %v", err)
	}

	// commit the root
	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing root: %v", err)
	}

	// Claim content 1
	cred := makeContentCredential(&didDoc.ID)
	_ = claims.AddProof(cred, didDoc.PublicKeys[0].ID, key)
	err = claimService.ClaimContent(cred)
	if err != nil {
		t.Errorf("problem creating content claim: %v", err)
	}

	proofBeforeCommit, err := claimService.GenerateProof(cred)
	if err != nil {
		t.Errorf("error generating proof: %v", err)
	}

	if proofBeforeCommit.BlockNumber != -1 {
		t.Errorf("block number should be -1 before committing the root")
	}

	// commit the root
	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing root: %v", err)
	}
	proof, err := claimService.GenerateProof(cred)
	if err != nil {
		t.Errorf("error generating proof: %v", err)
	}

	if proof.BlockNumber != 3 {
		t.Errorf("didn't get the right blocknumber")
	}

	existsProofBytes, err := hex.DecodeString(proof.ExistsInDIDMTProof)
	if err != nil {
		t.Errorf("couldn't decode exists proof: %v", err)
	}
	notRevokedBytes, err := hex.DecodeString(proof.NotRevokedInDIDMTProof)
	if err != nil {
		t.Errorf("couldn't decode nonrevoked proof: %v", err)
	}
	didRootBytes, err := hex.DecodeString(proof.DIDRootExistsProof)
	if err != nil {
		t.Errorf("couldn't decode didroot proof: %v", err)
	}

	existProof, err := merkletree.NewProofFromBytes(existsProofBytes)
	if err != nil {
		t.Errorf("couldn't build exists proof from bytes: %v", err)
	}
	notRevoked, err := merkletree.NewProofFromBytes(notRevokedBytes)
	if err != nil {
		t.Errorf("couldn't build not revoked proof from bytes: %v", err)
	}
	didRoot, err := merkletree.NewProofFromBytes(didRootBytes)
	if err != nil {
		t.Errorf("couldn't build did root proof from bytes: %v", err)
	}

	if err != nil {
		t.Errorf("couldn't get root of did tree: %v", err)
	}

	claimJSON, _ := json.Marshal(cred)
	hash := crypto.Keccak256(claimJSON)
	mhash, _ := multihash.EncodeName(hash, "keccak-256")
	hash34 := [34]byte{}
	copy(hash34[:], mhash)

	rdClaim, _ := claimtypes.NewClaimRegisteredDocument(hash34, signerDid, claimtypes.ContentCredentialDocType)

	entry := rdClaim.Entry()
	if !merkletree.VerifyProof(&proof.DIDRoot, existProof, entry.HIndex(), entry.HValue()) {
		t.Errorf("couldn't verify exists in did tree proof")
	}
	// revoked registered doc claim is always version 1
	rdClaim.Version = 1
	entryV1 := rdClaim.Entry()
	if !merkletree.VerifyProof(&proof.DIDRoot, notRevoked, entryV1.HIndex(), entryV1.HValue()) {
		t.Errorf("couldn't verify not revoked in proof")
	}

	rootClaim, _ := claimtypes.NewClaimSetRootKeyDID(signerDid, &proof.DIDRoot)
	rootClaim.Version = proof.DIDRootExistsVersion
	rootClaimEntry := rootClaim.Entry()

	if !merkletree.VerifyProof(&proof.Root, didRoot, rootClaimEntry.HIndex(), rootClaimEntry.HValue()) {
		t.Errorf("couldn't verify root tree proof")
	}

	err = claimService.RevokeClaim(cred, signerDid)
	if err != nil {
		t.Errorf("couldn't revoke claim")
	}

	_, err = claimService.GenerateProof(cred)
	if err == nil {
		t.Errorf("it should error if the claim is revoked")
	}

}

func TestClaimLicense(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didService, ethURI := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, _, err := makeService(db, didService, signedClaimStore)
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
	didDoc, err := ethuri.InitializeNewDocument(signerDid, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}
	subjDid, _ := ethuri.GenerateEthURIDID()

	license := makeLicenseCredential(signerDid, subjDid)
	err = claims.AddProof(license, didDoc.PublicKeys[0].ID, key)
	if err != nil {
		t.Errorf("error adding proof: %v", err)
	}
	err = claimService.ClaimLicense(license, signerDid)
	if err == nil {
		t.Errorf("should have errored because couldn't resolv the key")
	}
	err = claimService.CreateTreeForDIDWithPks(&didDoc.ID,
		[]*ecdsa.PublicKey{&key.PublicKey})
	if err != nil {
		t.Errorf("problem creating did tree: %v", err)
	}
	err = claimService.ClaimLicense(license, signerDid)
	if err != nil {
		t.Errorf("problem creating content claim: %v", err)
	}
	err = claimService.ClaimLicense(license, signerDid)
	if err == nil {
		t.Errorf("should err for duplicate claim")
	}
}
