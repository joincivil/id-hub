package claimsstore_test

import (
	"bytes"
	"testing"

	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

func TestVersionAPI(t *testing.T) {
	store, persister := pgStore(t)
	cleaner := testutils.DeleteCreatedEntities(persister.DB)
	defer cleaner()
	rootStore := store.WithPrefix(claimsstore.PrefixRootMerkleTree)

	rootMt, err := merkletree.NewMerkleTree(rootStore, 150)
	if err != nil {
		t.Errorf("couldn't create root merkle tree: %v", err)
	}

	dids := "did:ethuri:did1"
	did, err := didlib.Parse(dids)

	if err != nil {
		t.Errorf("couldn't parse did: %v", err)
	}

	root1 := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0c})

	root2 := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0c, 0x0d})

	root3 := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0e})

	root4 := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0c, 0x0c, 0x0b, 0x0b, 0x0c})

	root5 := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0xcc, 0x0b, 0x0b, 0x0b, 0x0c})

	claim1, err := claimtypes.NewClaimSetRootKeyDID(did, &root1)
	if err != nil {
		t.Errorf("couldn't create claim: %v", err)
	}
	claim2, err := claimtypes.NewClaimSetRootKeyDID(did, &root2)
	if err != nil {
		t.Errorf("couldn't create claim: %v", err)
	}
	claim3, err := claimtypes.NewClaimSetRootKeyDID(did, &root3)
	if err != nil {
		t.Errorf("couldn't create claim: %v", err)
	}
	claim4, err := claimtypes.NewClaimSetRootKeyDID(did, &root4)
	if err != nil {
		t.Errorf("couldn't create claim: %v", err)
	}
	claim5, err := claimtypes.NewClaimSetRootKeyDID(did, &root5)
	if err != nil {
		t.Errorf("couldn't create claim: %v", err)
	}
	rootBeforeAdding := rootMt.RootKey()
	err = rootMt.Add(claim1.Entry())
	if err != nil {
		t.Errorf("couldn't add claim to tree: %v", err)
	}
	// get next version of the claim
	version, err := persister.GetNextRootClaimVersion(did)
	if err != nil {
		t.Errorf("couldn't get next root claim version: %v", err)
	}
	if version != 1 {
		t.Errorf("unexpected version number: expected 1, got %v", version)
	}
	claim2.Version = version
	err = rootMt.Add(claim2.Entry())
	if err != nil {
		t.Errorf("couldn't add claim to tree: %v", err)
	}
	// get next version of the claim
	version, err = persister.GetNextRootClaimVersion(did)
	if err != nil {
		t.Errorf("couldn't get next root claim version: %v", err)
	}
	if version != 2 {
		t.Errorf("unexpected version number: expected 2, got %v", version)
	}
	claim3.Version = version
	err = rootMt.Add(claim3.Entry())
	if err != nil {
		t.Errorf("couldn't add claim to tree: %v", err)
	}
	rootAfterAdding3 := rootMt.RootKey()
	// get next version of the claim
	version, err = persister.GetNextRootClaimVersion(did)
	if err != nil {
		t.Errorf("couldn't get next root claim version: %v", err)
	}
	claim4.Version = version
	err = rootMt.Add(claim4.Entry())
	if err != nil {
		t.Errorf("couldn't add claim to tree: %v", err)
	}
	rootAfterAdding4 := rootMt.RootKey()
	// get next version of the claim
	version, err = persister.GetNextRootClaimVersion(did)
	if err != nil {
		t.Errorf("couldn't get next root claim version: %v", err)
	}
	claim5.Version = version
	err = rootMt.Add(claim5.Entry())
	if err != nil {
		t.Errorf("couldn't add claim to tree: %v", err)
	}

	snapShotAfter3, err := rootMt.Snapshot(rootAfterAdding3)
	if err != nil {
		t.Errorf("couldn't make snapshot: %v", err)
	}
	claim, err := persister.GetLatestRootClaimInSnapshot(did, snapShotAfter3)
	if err != nil {
		t.Errorf("error retrieving claim: %v", err)
	}
	if claim.Version != 2 {
		t.Errorf("wrong version on retrieved claim")
	}
	if !bytes.Equal(claim.RootKey[:], root3[:]) {
		t.Errorf("unexpected hash")
	}

	snapShotAfter4, err := rootMt.Snapshot(rootAfterAdding4)
	if err != nil {
		t.Errorf("error creating snapshot: %v", err)
	}
	claim, err = persister.GetLatestRootClaimInSnapshot(did, snapShotAfter4)
	if err != nil {
		t.Errorf("error retrieving claim: %v", err)
	}
	if claim.Version != 3 {
		t.Errorf("wrong version on retrieved claim")
	}
	if !bytes.Equal(claim.RootKey[:], root4[:]) {
		t.Errorf("unexpected hash")
	}

	snapShotBefore, err := rootMt.Snapshot(rootBeforeAdding)
	if err != nil {
		t.Errorf("error creating snapshot: %v", err)
	}
	_, err = persister.GetLatestRootClaimInSnapshot(did, snapShotBefore)
	if err != claimsstore.ErrNoRootCommitForDID {
		t.Errorf("should have got the no root commit for did error")
	}
}
