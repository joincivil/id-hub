package claimtypes_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

func TestFindLinkedDataProof(t *testing.T) {
	linkedData := &linkeddata.Proof{
		Type: string(linkeddata.SuiteTypeSecp256k1Signature),
	}

	proof, err := claimtypes.FindLinkedDataProof(linkedData)
	if err != nil || proof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		t.Errorf("should work for a pointer to a linked data proof")
	}

	proof, err = claimtypes.FindLinkedDataProof(*linkedData)
	if err != nil || proof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		t.Errorf("should work for a linked data proof")
	}

	linkedDataMap := make(map[string]interface{})
	linkedDataMap["type"] = string(linkeddata.SuiteTypeSecp256k1Signature)
	proof, err = claimtypes.FindLinkedDataProof(linkedDataMap)
	if err != nil || proof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		t.Errorf("should work for a map with the correct fields")
	}

	badLinkedDataMap := make(map[string]interface{})
	badLinkedDataMap["someprop"] = "somestring"
	_, err = claimtypes.FindLinkedDataProof(badLinkedDataMap)
	if err == nil {
		t.Errorf("shouldnt work for a map with incorrect fields")
	}

	slice1 := []interface{}{linkedData}
	proof, err = claimtypes.FindLinkedDataProof(slice1)
	if err != nil || proof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		t.Errorf("should work for a slice with a pointer to linked data proof")
	}

	slice2 := []interface{}{*linkedData}
	proof, err = claimtypes.FindLinkedDataProof(slice2)
	if err != nil || proof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		t.Errorf("should work for a slice with a linked data proof")
	}

	slice3 := []interface{}{linkedDataMap}
	proof, err = claimtypes.FindLinkedDataProof(slice3)
	if err != nil || proof.Type != string(linkeddata.SuiteTypeSecp256k1Signature) {
		t.Errorf("should work for a slice with a map")
	}
}
