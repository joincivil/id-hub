package claims

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	icore "github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
	isrv "github.com/iden3/go-iden3-core/services/claimsrv"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	didlib "github.com/ockam-network/did"
)

// Service is a service for creating and reading claims
type Service struct {
	rootMt           *merkletree.MerkleTree
	treeStore        db.Storage
	signedClaimStore *claimsstore.SignedClaimPGPersister
	didService       *did.Service
}

// NewService returns a new service
func NewService(treeStore db.Storage, signedClaimStore *claimsstore.SignedClaimPGPersister, didService *did.Service) (*Service, error) {
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
	}, nil
}

func (s *Service) addNewRootClaim(didMt *merkletree.MerkleTree, userDid *didlib.DID) error {
	claimSetRootKey, err := NewClaimSetRootKeyDID(userDid, didMt.RootKey())
	if err != nil {
		return err
	}
	// get next version of the claim
	version, err := isrv.GetNextVersion(s.rootMt, claimSetRootKey.Entry().HIndex())
	if err != nil {
		return err
	}
	claimSetRootKey.Version = version
	err = s.rootMt.Add(claimSetRootKey.Entry())
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) buildDIDMt(userDid *didlib.DID) (*merkletree.MerkleTree, error) {
	bid, err := claimsstore.DIDToBinary(userDid)
	if err != nil {
		return nil, err
	}
	didStore := s.treeStore.WithPrefix(bid)
	return merkletree.NewMerkleTree(didStore, 150)
}

// CreateTreeForDID creates a new merkle tree for the did and registers a public key
func (s *Service) CreateTreeForDID(userDid *didlib.DID, signPk *ecdsa.PublicKey) error {
	didMt, err := s.buildDIDMt(userDid)
	if err != nil {
		return err
	}
	claimKey := icore.NewClaimAuthorizeKSignSecp256k1(signPk)
	err = didMt.Add(claimKey.Entry())
	if err != nil {
		return err
	}
	return s.addNewRootClaim(didMt, userDid)
}

func (s *Service) verifyCredential(cred *claimsstore.ContentCredential, userMt *merkletree.MerkleTree, signerDid *didlib.DID) (bool, error) {
	if cred.Proof.Type != string(did.LDSuiteTypeSecp256k1Signature) {
		return false, errors.New("Only Secp256k1 signature types are implemented")
	}
	pubkey, err := s.didService.GetKeyFromDIDDocument(signerDid)
	if err != nil {
		return false, err
	}
	if pubkey.Type != did.LDSuiteTypeSecp256k1Verification {
		return false, errors.New("Only secp256k1 signatures are currently supported")
	}
	pubbytes, err := hex.DecodeString(*pubkey.PublicKeyHex)
	if err != nil {
		return false, err
	}
	ecpub, err := crypto.UnmarshalPubkey(pubbytes[:])
	if err != nil {
		return false, err
	}

	if !isrv.CheckKSignInIddb(userMt, ecpub) {
		return false, errors.New("key used to sign has not been claimed in the merkle tree")
	}

	canoncred, err := CanonicalizeCredential(cred)
	if err != nil {
		return false, err
	}
	sigbytes, err := hex.DecodeString(cred.Proof.ProofValue)
	if err != nil {
		return false, err
	}
	verified := crypto.VerifySignature(pubbytes, crypto.Keccak256(canoncred), sigbytes)
	return verified, nil
}

// ClaimContent takes a content credential and saves it to the signed credential table
// and then registers it in the tree
func (s *Service) ClaimContent(cred *claimsstore.ContentCredential) error {
	signerDid, err := didlib.Parse(cred.Proof.Creator)
	if err != nil {
		return err
	}

	// for a content claim the signer should also be the issuer and holder
	didMt, err := s.buildDIDMt(signerDid)
	if err != nil {
		return err
	}
	verified, err := s.verifyCredential(cred, didMt, signerDid)
	if err != nil {
		return err
	}
	if !verified {
		return errors.New("could not verify string on credential")
	}
	hash, err := s.signedClaimStore.AddCredential(cred)
	if err != nil {
		return err
	}
	hashb, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}
	if len(hashb) > 32 {
		return errors.New("hash hex string is the wrong size")
	}
	hashb32 := [32]byte{}
	copy(hashb32[:], hashb)

	claim, err := NewClaimRegisteredDocument(hashb32, signerDid, ContentCredentialDocType)
	if err != nil {
		return err
	}
	err = didMt.Add(claim.Entry())
	if err != nil {
		return err
	}

	return s.addNewRootClaim(didMt, signerDid)
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
		claim, err := NewClaimFromEntry(entry)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}
	return claims, nil
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
