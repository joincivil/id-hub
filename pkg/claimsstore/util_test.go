package claimsstore_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/claimsstore"
	didlib "github.com/ockam-network/did"
)

func TestDIDToBinary(t *testing.T) {
	dids := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	did, err := didlib.Parse(dids)
	if err != nil {
		t.Errorf("problem creating DID: %v", err)
	}
	didb, err := claimsstore.DIDToBinary(did)
	if err != nil {
		t.Errorf("could not convert did to binary: %v", err)
	}
	if len(didb) != 32 {
		t.Errorf("did binary is the wrong size")
	}
	reclaimedDid, err := claimsstore.BinaryToDID(didb)
	if err != nil {
		t.Errorf("problem turning binary back to did: %v", err)
	}

	if reclaimedDid.String() != dids {
		t.Errorf("did came back mutated")
	}

	dids = "did:superduperlongdidmethodname:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"

	did, err = didlib.Parse(dids)
	if err != nil {
		t.Errorf("problem creating DID: %v", err)
	}
	_, err = claimsstore.DIDToBinary(did)

	if err != claimsstore.ErrTooLongDIDMethod {
		t.Errorf("should have errored for super long method name")
	}

}
