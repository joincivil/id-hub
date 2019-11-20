package utils_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/utils"
)

func TestStrOrEmptyStr(t *testing.T) {
	res := utils.StrOrEmptyStr(nil)
	if res != "" {
		t.Errorf("Should have returned an empty string")
	}

	testStr := "thisisatest"
	res = utils.StrOrEmptyStr(&testStr)
	if res == "" {
		t.Errorf("Should not have returned an empty string")
	}
	if res != testStr {
		t.Errorf("Should have returned the test string")
	}

	testStr = ""
	res = utils.StrOrEmptyStr(&testStr)
	if res != "" {
		t.Errorf("Should have returned an empty string")
	}
}

func TestStrToPtr(t *testing.T) {
	testStr := "thisisatest"
	strPtr := utils.StrToPtr(testStr)
	if *strPtr != testStr {
		t.Errorf("Should have returned the test string")
	}
}
