package claimtypes_test

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	didlib "github.com/ockam-network/did"
)

func TestClaimRegisteredDoc(t *testing.T) {
	dids := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	did, err := didlib.Parse(dids)
	if err != nil {
		t.Errorf("failed to parse did: %v", err)
	}
	hash := crypto.Keccak256([]byte("abcdefg"))
	hash32 := [32]byte{}
	copy(hash32[:], hash)
	claim, err := claimtypes.NewClaimRegisteredDocument(hash32, did, 0)
	if err != nil {
		t.Errorf("error making claim: %v", err)
	}
	entry := claim.Entry()
	claim2 := claimtypes.NewClaimRegisteredDocumentFromEntry(entry)
	if !bytes.Equal(claim.ContentHash[:], claim2.ContentHash[:]) {
		t.Errorf("couldn't successfully recover claim from entry")
	}
	hashDid, err := claimtypes.HashDID(did)
	if err != nil {
		t.Errorf("Couldn't hash the did: %v", err)
	}
	if !bytes.Equal(claim.DID[:], hashDid) {
		t.Errorf("recoverd DID did not match expected hash value")
	}
}
