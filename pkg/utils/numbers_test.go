package utils_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/utils"
)

func TestIntOrEmptyInt(t *testing.T) {
	res := utils.IntOrEmptyInt(nil)
	if res != 0 {
		t.Errorf("Should have returned an empty int")
	}

	testInt := 100
	res = utils.IntOrEmptyInt(&testInt)
	if res == 0 {
		t.Errorf("Should not have returned an empty int")
	}
	if res != testInt {
		t.Errorf("Should have returned the test string")
	}

	testInt = 0
	res = utils.IntOrEmptyInt(&testInt)
	if res != 0 {
		t.Errorf("Should have returned an empty int")
	}
}

func TestIntToPtr(t *testing.T) {
	testInt := 100
	strPtr := utils.IntToPtr(testInt)
	if *strPtr != testInt {
		t.Errorf("Should have returned the test int")
	}
}
