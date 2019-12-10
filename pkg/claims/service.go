package claims

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"

	log "github.com/golang/glog"

	"github.com/ethereum/go-ethereum/crypto"

	icore "github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/merkletree"
	isrv "github.com/iden3/go-iden3-core/services/claimsrv"

	"github.com/joincivil/go-common/pkg/lock"
	"github.com/joincivil/go-common/pkg/numbers"

	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"

	didlib "github.com/ockam-network/did"
)

var (
	lockExpirationMillis = numbers.IntToPtr(1000 * 5)
)

// Service is a service for creating and reading claims
type Service struct {
	rootMt           *merkletree.MerkleTree
	treeStore        *claimsstore.PGStore
	signedClaimStore *claimsstore.SignedClaimPGPersister
	didService       *did.Service
	rootService      *RootService
	dlock            lock.DLock
}

// NewService returns a new service
func NewService(treeStore *claimsstore.PGStore, signedClaimStore *claimsstore.SignedClaimPGPersister,
	didService *did.Service, rootService *RootService, dlock lock.DLock) (*Service, error) {
	rootStore := treeStore.WithPrefix(claimsstore.PrefixRootMerkleTree)

	rootMt, err := merkletree.NewMerkleTree(rootStore, 150)
	if err != nil {
		return nil, err
	}

	return &Service{
		rootMt:           rootMt,
		treeStore:        treeStore,
		signedClaimStore: signedClaimStore,
		didService:       didService,
		rootService:      rootService,
		dlock:            dlock,
	}, nil
}

func (s *Service) getSignerDID(proofs interface{}) (*didlib.DID, error) {
	linkedDataProof, err := claimtypes.FindLinkedDataProof(proofs)
	if err != nil {
		return nil, errors.Wrap(err, "getSignerDID.FindLinkedDataProof")
	}
	return didlib.Parse(linkedDataProof.Creator)
}

func (s *Service) generateProofAndNonRevokeFromEntry(entry *merkletree.Entry, tree *merkletree.MerkleTree) (*merkletree.Proof, *merkletree.Proof, error) {
	hi := entry.HIndex()
	proof, err := tree.GenerateProof(hi, tree.RootKey())
	if err != nil {
		return &merkletree.Proof{}, &merkletree.Proof{}, errors.Wrap(err, "generateProofAndNonRevokeFromEntry.tree.GenerateProof")
	}
	leafData, err := tree.GetDataByIndex(hi)
	if err != nil {
		return &merkletree.Proof{}, &merkletree.Proof{}, errors.Wrap(err, "generateProofAndNonRevokeFromEntry.tree.GetDataByIndex")
	}
	revoke, err := icore.GetNonRevocationMTProof(tree, leafData, hi)
	if err != nil {
		return &merkletree.Proof{}, &merkletree.Proof{}, errors.Wrap(err, "generateProofAndNonRevokeFromEntry.GetNonRevocationMTProof")
	}
	return proof, revoke, nil
}

func (s *Service) getLastRootClaim(claim *claimtypes.ContentCredential,
	rootSnapShotTree *merkletree.MerkleTree) (*claimtypes.ClaimSetRootKeyDID, *merkletree.MerkleTree, error) {
	signerDID, err := s.getSignerDID(claim.Proof)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getLastRootClaim.getSignerDID")
	}

	lastRootClaimForDID, err := s.treeStore.NodePersister.GetLatestRootClaimInSnapshot(signerDID, rootSnapShotTree)
	if err != nil {
		return nil, nil, errors.Wrap(err,
			"getLastRootClaim NodePersister.GetLatestRootClaimInSnapshot failed to get last root for did")
	}

	didMT, err := s.buildDIDMt(signerDID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getLastRootClaim.buildDIDMt")
	}

	didTreeSnapshot, err := didMT.Snapshot(&lastRootClaimForDID.RootKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getLastRootClaim.didMT.Snapshot")
	}
	return lastRootClaimForDID, didTreeSnapshot, nil
}

func (s *Service) getRootSnapshot(commit *claimsstore.RootCommit) (*merkletree.MerkleTree, error) {
	lastRoot, err := hex.DecodeString(commit.Root[2:])
	if err != nil {
		return nil, errors.Wrap(err, "getRootSnapshot hex.DecodeString")
	}
	if len(lastRoot) != 32 {
		return nil, errors.New("root hash should be 32 bytes")
	}

	lastrootHash := merkletree.Hash{}
	copy(lastrootHash[:], lastRoot)

	return s.rootMt.Snapshot(&lastrootHash)
}

func (s *Service) makeContentClaimFromCred(claim claimtypes.Credential, claimer *didlib.DID) (*claimtypes.ClaimRegisteredDocument, error) {
	claimJSON, err := json.Marshal(claim)
	if err != nil {
		return nil, errors.Wrap(err, "makeContentClaimFromCred json.Marshal")
	}
	hash := crypto.Keccak256(claimJSON)
	mhash, err := multihash.EncodeName(hash, "keccak-256")
	if err != nil {
		return nil, errors.Wrap(err, "makeContentClaimFromCred multihash.EncodeName")
	}
	hash34 := [34]byte{}
	copy(hash34[:], mhash)
	return claimtypes.NewClaimRegisteredDocument(hash34, claimer, claimtypes.ContentCredentialDocType)
}

// GenerateProof returns a proof that the content credential is in the tree and on the blockchain
func (s *Service) GenerateProof(claim *claimtypes.ContentCredential, claimer *didlib.DID) (*MTProof, error) {
	lastRootCommit, err := s.rootService.GetLatest()
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.rootService.GetLatest")
	}

	lastRootSnapshot, err := s.getRootSnapshot(lastRootCommit)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.getRootSnapshot")
	}

	latestRootClaim, didTreeSnapshot, err := s.getLastRootClaim(claim, lastRootSnapshot)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.getLastRootClaim")
	}

	rdClaim, err := s.makeContentClaimFromCred(claim, claimer)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.makeContentClaimFromCred")
	}

	entry := rdClaim.Entry()
	existsInDIDMTProof, notRevokedInDIDMTProof, err := s.generateProofAndNonRevokeFromEntry(entry, didTreeSnapshot)
	// will error here if the proof is revoked in the anchored tree
	if err != nil && strings.HasSuffix(err.Error(), "node index not found in the DB") {
		existsInDIDMTProof = &merkletree.Proof{
			Existence: false,
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.generateProofAndNonRevokeFromEntry")
	}

	didMt, err := s.buildDIDMt(claimer)
	if err != nil {
		return nil, err
	}

	existsInDIDMTProofNoAnchor, notRevokedInDIDMTProofNoAnchor, err := s.generateProofAndNonRevokeFromEntry(entry, didMt)
	// will anchore here if the the proof is revoked in the unanchored tree
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.generateProofAndNonRevokeFromEntry")
	}

	// Will error before this for revocation
	// if the claim exists in the anchored tree
	if existsInDIDMTProof.Existence {
		didRootExistsProof, err := lastRootSnapshot.GenerateProof(latestRootClaim.Entry().HIndex(), lastRootSnapshot.RootKey())
		if err != nil {
			return nil, errors.Wrap(err, "GenerateProof.lastRootSnapshot.GenerateProof")
		}

		return &MTProof{
			ExistsInDIDMTProof:     hex.EncodeToString(existsInDIDMTProof.Bytes()),
			NotRevokedInDIDMTProof: hex.EncodeToString(notRevokedInDIDMTProof.Bytes()),
			DIDRootExistsProof:     hex.EncodeToString(didRootExistsProof.Bytes()),
			DIDRootExistsVersion:   latestRootClaim.Version,
			BlockNumber:            lastRootCommit.BlockNumber,
			ContractAddress:        common.HexToAddress(lastRootCommit.ContractAddress),
			TXHash:                 common.HexToHash(lastRootCommit.TransactionHash),
			Root:                   *lastRootSnapshot.RootKey(),
			DIDRoot:                *didTreeSnapshot.RootKey(),
			CommitterAddress:       common.HexToAddress(lastRootCommit.CommitterAddress),
			DID:                    claimer.String(),
		}, nil
	}

	rootClaimNoAnchor, _, err := s.getLastRootClaim(claim, s.rootMt)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.getLastRootClaim failed for current root tree")
	}

	didRootExistsProofNoAnchor, err := lastRootSnapshot.GenerateProof(rootClaimNoAnchor.Entry().HIndex(), s.rootMt.RootKey())
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof.lastRootSnapshot.GenerateProof")
	}

	return &MTProof{
		ExistsInDIDMTProof:     hex.EncodeToString(existsInDIDMTProofNoAnchor.Bytes()),
		NotRevokedInDIDMTProof: hex.EncodeToString(notRevokedInDIDMTProofNoAnchor.Bytes()),
		DIDRootExistsProof:     hex.EncodeToString(didRootExistsProofNoAnchor.Bytes()),
		DIDRootExistsVersion:   rootClaimNoAnchor.Version,
		BlockNumber:            -1,
		ContractAddress:        common.HexToAddress(lastRootCommit.ContractAddress),
		TXHash:                 common.HexToHash("0x0"),
		Root:                   *s.rootMt.RootKey(),
		DIDRoot:                *didMt.RootKey(),
		CommitterAddress:       common.HexToAddress(lastRootCommit.CommitterAddress),
		DID:                    claimer.String(),
	}, nil

}

func (s *Service) addNewRootClaim(userDid *didlib.DID) error {
	didMt, err := s.buildDIDMt(userDid)
	if err != nil {
		return err
	}
	claimSetRootKey, err := claimtypes.NewClaimSetRootKeyDID(userDid, didMt.RootKey())
	if err != nil {
		return errors.Wrap(err, "addNewRootClaim.NewClaimSetRootKeyDID")
	}

	// distributed locking around the root claim version and addition
	lerr := s.dlock.Lock(userDid.String(), lockExpirationMillis)
	if lerr != nil {
		return errors.Wrapf(lerr, "addNewRootClaim.Lock: %v", userDid.String())
	}

	// get next version of the claim
	version, err := s.treeStore.NodePersister.GetNextRootClaimVersion(userDid)
	if gorm.IsRecordNotFoundError(err) {
		version = 0
	} else if err != nil {
		return errors.Wrap(err, "addNewRootClaim.NodePersister.GetNextRootClaimVersion")
	}

	claimSetRootKey.Version = version
	err = s.rootMt.Add(claimSetRootKey.Entry())
	lerr = s.dlock.Unlock(userDid.String())
	if lerr != nil {
		log.Infof("addNewRootClaim.Unlock: err: %v, did: %v", lerr, userDid.String())
	}
	if err != nil {
		return errors.Wrap(err, "addNewRootClaim.rootMt.Add")
	}

	return nil
}

func (s *Service) buildDIDMt(userDid *didlib.DID) (*merkletree.MerkleTree, error) {
	didStringOnlyMethodID := did.MethodIDOnly(userDid)
	bid := []byte(didStringOnlyMethodID)
	didStore := s.treeStore.WithPrefix(bid)
	return merkletree.NewMerkleTree(didStore, 150)
}

// CreateTreeForDIDWithPks creates a new merkle tree for the did and
// registers a slice of public key that can be used for signing with this did
// Can also be used to add additional key claims to the userDID MT
func (s *Service) CreateTreeForDIDWithPks(userDid *didlib.DID, signPks []*ecdsa.PublicKey) error {
	if len(signPks) == 0 {
		return errors.New("at least one public key required")
	}

	didMt, err := s.buildDIDMt(userDid)
	if err != nil {
		return err
	}

	// Claim all the valid public keys that could be used to sign
	var claimKey *icore.ClaimAuthorizeKSignSecp256k1
	var addRoot bool
	for _, k := range signPks {
		// Check to ensure the key claim isn't already in tree
		if isrv.CheckKSignInIddb(didMt, k) {
			continue
		}

		claimKey = icore.NewClaimAuthorizeKSignSecp256k1(k)
		err = didMt.Add(claimKey.Entry())
		if err != nil {
			return errors.Wrap(err, "unable to add signing key claim")
		}
		addRoot = true
	}

	if addRoot {
		return s.addNewRootClaim(userDid)
	}

	return nil
}

// CreateTreeForDID creates a new tree for a user DID if it does not exist already.
func (s *Service) CreateTreeForDID(userDid *didlib.DID) error {
	doc, err := s.didService.GetDocumentFromDID(userDid)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve document for did")
	}
	if doc == nil {
		return errors.New("no doc found for did")
	}

	return s.CreateTreeForDIDWithPks(
		userDid,
		did.DocPublicKeyToEcdsaKeys(doc.PublicKeys),
	)
}

func (s *Service) verifyCredential(cred claimtypes.Credential) (bool, error) {
	linkedDataProof, err := cred.FindLinkedDataProof()
	if err != nil {
		return false, errors.Wrap(err, "verifyCredential.FindLinkedDataProof")
	}
	if linkedDataProof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		return false, errors.New("Only Secp256k1 signature types are implemented")
	}
	signerDid, err := didlib.Parse(linkedDataProof.Creator)
	if err != nil {
		return false, errors.Wrap(err, "verifyCredential parse creator did")
	}
	pubkey, err := s.didService.GetKeyFromDIDDocument(signerDid)
	if err != nil {
		return false, err
	}
	ecpub, err := pubkey.AsEcdsaPubKey()
	if err != nil {
		return false, err
	}

	signerMt, err := s.buildDIDMt(signerDid)
	if err != nil {
		return false, errors.Wrap(err, "verifyCredential.buildDIDMt")
	}

	if !isrv.CheckKSignInIddb(signerMt, ecpub) {
		return false, errors.New("key used to sign has not been claimed in the merkle tree")
	}
	canoncred, err := cred.CanonicalizeCredential()
	if err != nil {
		return false, err
	}
	sigbytes, err := hex.DecodeString(linkedDataProof.ProofValue)
	if err != nil {
		return false, err
	}
	recoveredPubkey, err := crypto.SigToPub(crypto.Keccak256(canoncred), sigbytes)
	if err != nil {
		return false, err
	}
	recoveredBytes := crypto.FromECDSAPub(recoveredPubkey)
	pubBytes := crypto.FromECDSAPub(ecpub)
	return bytes.Equal(recoveredBytes, pubBytes), nil
}

// ClaimContent takes a content credential and saves it to the signed credential table
// and then registers it in the tree
func (s *Service) ClaimContent(cred *claimtypes.ContentCredential) error {
	signerDid, err := s.getSignerDID(cred.Proof)
	if err != nil {
		return errors.Wrap(err, "ClaimContent.getSignerDID")
	}

	if signerDid.Fragment == "" {
		return errors.New("claimcontent expecting fragment on did for proof creator")
	}

	// for a content claim the signer should also be the issuer and holder
	didMt, err := s.buildDIDMt(signerDid)
	if err != nil {
		return errors.Wrap(err, "claimcontent.builddidMt")
	}
	verified, err := s.verifyCredential(cred)
	if err != nil {
		return errors.Wrap(err, "claimcontent.verifycredential")
	}
	if !verified {
		return errors.New("could not verify credential")
	}
	hash, err := s.signedClaimStore.AddCredential(cred)
	if err != nil {
		return errors.Wrap(err, "claimcontent.addcredential")
	}
	hashb, err := hex.DecodeString(hash)
	if err != nil {
		return errors.Wrap(err, "claimcontent.decodestring")
	}
	if len(hashb) > 34 {
		return errors.New("hash hex string is the wrong size")
	}
	hashb34 := [34]byte{}
	copy(hashb34[:], hashb)

	claim, err := claimtypes.NewClaimRegisteredDocument(hashb34, signerDid, claimtypes.ContentCredentialDocType)
	if err != nil {
		return errors.Wrap(err, "claimcontent.newclaimregistereddocument")
	}
	err = didMt.Add(claim.Entry())
	if err != nil {
		return errors.Wrap(err, "claimcontent.add")
	}
	err = s.addNewRootClaim(signerDid)
	if err != nil {
		return errors.Wrap(err, "claimcontent.addnewrootclaim")
	}

	return nil
}

// RevokeClaim adds a revocation to the registered doc associated with a credential
func (s *Service) RevokeClaim(cred claimtypes.Credential, claimer *didlib.DID) error {
	didMt, err := s.buildDIDMt(claimer)
	if err != nil {
		return errors.Wrap(err, "revokeclaim.builddidMt")
	}

	rdClaim, err := s.makeContentClaimFromCred(cred, claimer)
	if err != nil {
		return errors.Wrap(err, "RevokeClaim.makeContentClaimFromCred")
	}

	rdClaim.Version = 1 // 1 signifies revokation for all registered document claims

	err = didMt.Add(rdClaim.Entry())
	if err != nil {
		return errors.Wrap(err, "RevokeClaim.add")
	}

	err = s.addNewRootClaim(claimer)
	if err != nil {
		return errors.Wrap(err, "RevokeClaim.addnewrootclaim")
	}

	return nil
}

// ClaimLicense adds a license claim to the claimers claim tree
func (s *Service) ClaimLicense(cred *claimtypes.LicenseCredential, claimer *didlib.DID) error {
	didMt, err := s.buildDIDMt(claimer)
	if err != nil {
		return errors.Wrap(err, "claimlicense.builddidMt")
	}

	verified, err := s.verifyCredential(cred)
	if err != nil {
		return errors.Wrap(err, "ClaimLicense.verifycredential")
	}

	if !verified {
		return errors.New("could not verify credential")
	}

	hash, err := s.signedClaimStore.AddCredential(cred)
	if err != nil {
		return errors.Wrap(err, "ClaimLicense.addcredential")
	}
	hashb, err := hex.DecodeString(hash)
	if err != nil {
		return errors.Wrap(err, "ClaimLicense.decodestring")
	}
	if len(hashb) > 34 {
		return errors.New("hash hex string is the wrong size")
	}
	hashb34 := [34]byte{}
	copy(hashb34[:], hashb)

	claim, err := claimtypes.NewClaimRegisteredDocument(hashb34, claimer, claimtypes.LicenseCredentialDocType)
	if err != nil {
		return errors.Wrap(err, "ClaimLicense.newclaimregistereddocument")
	}
	err = didMt.Add(claim.Entry())
	if err != nil {
		return errors.Wrap(err, "ClaimLicense.add")
	}
	err = s.addNewRootClaim(claimer)
	if err != nil {
		return errors.Wrap(err, "ClaimLicense.addnewrootclaim")
	}

	return nil
}

func getClaimsForTree(tree *merkletree.MerkleTree) ([]merkletree.Claim, error) {
	rootKey := tree.RootKey()

	entries, err := tree.DumpClaims(rootKey)
	if err != nil {
		return nil, err
	}
	claims := []merkletree.Claim{}
	for _, v := range entries {
		entryb, err := hex.DecodeString(v[2:])
		if err != nil {
			return nil, err
		}
		entry, err := merkletree.NewEntryFromBytes(entryb)
		if err != nil {
			return nil, err
		}
		claim, err := claimtypes.NewClaimFromEntry(entry)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}
	return claims, nil
}

// ClaimsToContentCredentials converts a list of merkletree.Claim interfaces
// to concrete ContentCredentials. Filters out claims not of type
// ContentCredential.
func (s *Service) ClaimsToContentCredentials(clms []merkletree.Claim) (
	[]*claimtypes.ContentCredential, error) {
	creds := make([]*claimtypes.ContentCredential, 0, len(clms))

	for _, v := range clms {
		switch tv := v.(type) {
		case claimtypes.ClaimRegisteredDocument, *claimtypes.ClaimRegisteredDocument:
			// XXX(PN): These are coming in as both value and by reference, normal?
			var regDoc claimtypes.ClaimRegisteredDocument
			d, ok := tv.(*claimtypes.ClaimRegisteredDocument)
			if ok {
				regDoc = *d
			} else {
				regDoc = tv.(claimtypes.ClaimRegisteredDocument)
			}

			if regDoc.DocType == claimtypes.ContentCredentialDocType {
				claimHash := hex.EncodeToString(regDoc.ContentHash[:])
				// XXX(PN): Needs a bulk loader here
				signed, err := s.signedClaimStore.GetCredentialByMultihash(claimHash)
				if err != nil {
					return nil, errors.Wrapf(err, "could not retrieve credential: hash: %v, err: %v", claimHash, err)
				}

				signedCont, ok := signed.(*claimtypes.ContentCredential)
				if ok {
					creds = append(creds, signedCont)
				}
			}

		case *icore.ClaimAuthorizeKSignSecp256k1:
			// Known claim type to ignore here

		default:
			log.Errorf("Unknown claim type, is %T", v)
		}
	}

	return creds, nil
}

// GetMerkleTreeClaimsForDid returns all the claims in a DID's merkletree
func (s *Service) GetMerkleTreeClaimsForDid(userDid *didlib.DID) ([]merkletree.Claim, error) {
	didMt, err := s.buildDIDMt(userDid)
	if err != nil {
		return nil, err
	}
	return getClaimsForTree(didMt)
}

// GetRootMerkleTreeClaims returns all root claims
func (s *Service) GetRootMerkleTreeClaims() ([]merkletree.Claim, error) {
	return getClaimsForTree(s.rootMt)
}

// GetDIDRoot returns the root hash of a dids tree
func (s *Service) GetDIDRoot(did *didlib.DID) (*merkletree.Hash, error) {
	didMt, err := s.buildDIDMt(did)
	if err != nil {
		return nil, err
	}
	return didMt.RootKey(), nil
}
