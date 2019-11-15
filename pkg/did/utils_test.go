package did_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/did"

	didlib "github.com/ockam-network/did"
)

func TestMethodIDOnlyFromSTring(t *testing.T) {
	testDidMethodID := "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485"
	testDid := "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1"

	onlyMethodID, err := did.MethodIDOnlyFromString(testDid)
	if err != nil {
		t.Fatal("should have not gotten error")
	}
	if onlyMethodID != testDidMethodID {
		t.Error("should have gotten proper did without fragments/paths")
	}

	onlyMethodID, err = did.MethodIDOnlyFromString(testDidMethodID)
	if err != nil {
		t.Fatal("should have not gotten error")
	}
	if onlyMethodID != testDidMethodID {
		t.Error("should have gotten proper did without fragments/paths")
	}
}

func TestMethodIDOnly(t *testing.T) {
	testDidMethodID := "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485"
	testDid := "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1"

	d, _ := didlib.Parse(testDid)
	onlyMethodID := did.MethodIDOnly(d)
	if onlyMethodID != testDidMethodID {
		t.Error("should have gotten proper did without fragments/paths")
	}

	d, _ = didlib.Parse(testDidMethodID)
	onlyMethodID = did.MethodIDOnly(d)
	if onlyMethodID != testDidMethodID {
		t.Error("should have gotten proper did without fragments/paths")
	}
}

func TestCopyDID(t *testing.T) {
	d, _ := did.GenerateEthURIDID()
	cpy := did.CopyDID(d)
	if cpy.String() != d.String() {
		t.Errorf("Should have matching DID strings")
	}
	if cpy == d {
		t.Errorf("Should not be the same exact struct value in mem")
	}
}

func TestValidDid(t *testing.T) {
	if did.ValidDid("notavaliddid") {
		t.Errorf("Should not have returned true as valid did")
	}
	if did.ValidDid("") {
		t.Errorf("Should not have returned true as valid did")
	}
	if did.ValidDid("uri:123345") {
		t.Errorf("Should not have returned true as valid did")
	}
	if !did.ValidDid("did:uri:123345") {
		t.Errorf("Should have returned true as valid did")
	}
}
