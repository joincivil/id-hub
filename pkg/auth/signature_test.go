package auth_test

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	ceth "github.com/joincivil/go-common/pkg/eth"
	ctime "github.com/joincivil/go-common/pkg/time"
	"github.com/joincivil/id-hub/pkg/auth"
)

func TestSignVerifyEcdsaRequestSignature(t *testing.T) {
	// Create a test signature for request message
	acct, _ := ceth.MakeAccount()
	pk := acct.Key

	did := "did:ethurl:123456"
	reqTs := ctime.CurrentEpochSecsInInt()
	signature, err := auth.SignEcdsaRequestMessage(pk, did, reqTs)
	if err != nil {
		t.Fatalf("Should not have failed to sign message")
	}
	if signature == "" {
		t.Fatalf("Should have returned a non-empty signature")
	}

	// Verify this signature is valid
	pubKeyBys := crypto.FromECDSAPub(&pk.PublicKey)
	pubKeyHex := hex.EncodeToString(pubKeyBys)
	verified, err := auth.VerifyEcdsaRequestSignature(pubKeyHex, signature, did,
		reqTs, auth.DefaultRequestGracePeriodSecs)
	if err != nil {
		t.Errorf("Should not have returned error when verifying: err: %v", err)
	}
	if !verified {
		t.Errorf("Should have verified this signature")
	}
}

func TestSignVerifyEcdsaRequestSignatureGracePeriod(t *testing.T) {
	acct, _ := ceth.MakeAccount()
	pk := acct.Key
	did := "did:ethurl:123456"

	// Out of grace period range, it's old
	reqTs := ctime.CurrentEpochSecsInInt() - 60*30
	signature, err := auth.SignEcdsaRequestMessage(pk, did, reqTs)
	if err != nil {
		t.Fatalf("Should not have failed to sign message")
	}
	if signature == "" {
		t.Fatalf("Should have returned a non-empty signature")
	}

	pubKeyBys := crypto.FromECDSAPub(&pk.PublicKey)
	pubKeyHex := hex.EncodeToString(pubKeyBys)
	verified, err := auth.VerifyEcdsaRequestSignature(pubKeyHex, signature, did,
		reqTs, auth.DefaultRequestGracePeriodSecs)
	if err != nil {
		t.Errorf("Should not have returned error when verifying: err: %v", err)
	}
	if verified {
		t.Errorf("Should have not verified this signature")
	}

	// Out of grace period range, it's in the future
	reqTs = ctime.CurrentEpochSecsInInt() + 60*30
	signature, err = auth.SignEcdsaRequestMessage(pk, did, reqTs)
	if err != nil {
		t.Fatalf("Should not have failed to sign message")
	}
	if signature == "" {
		t.Fatalf("Should have returned a non-empty signature")
	}

	verified, err = auth.VerifyEcdsaRequestSignature(pubKeyHex, signature, did, reqTs,
		auth.DefaultRequestGracePeriodSecs)
	if err != nil {
		t.Errorf("Should not have returned error when verifying")
	}
	if verified {
		t.Errorf("Should have not verified this signature")
	}
}

func TestSignVerifyEcdsaRequestSignatureErrs(t *testing.T) {
	acct, _ := ceth.MakeAccount()
	pk := acct.Key

	// Bad did when signing
	did := "did"
	reqTs := ctime.CurrentEpochSecsInInt()
	signature, err := auth.SignEcdsaRequestMessage(pk, did, reqTs)
	if err == nil {
		t.Fatalf("Should have failed to sign message")
	}
	if signature != "" {
		t.Fatalf("Should have returned an empty signature")
	}

	// Bad publicKey when verifying
	did = "did:ethurl:123456"
	badPubKey := "thisisabadkey"
	verified, err := auth.VerifyEcdsaRequestSignature(badPubKey, signature, did, reqTs,
		auth.DefaultRequestGracePeriodSecs)
	if err == nil {
		t.Errorf("Should have returned error when verifying")
	}
	if verified {
		t.Errorf("Should have not verified this signature")
	}

	// Bad signature when verifying
	pubKeyBys := crypto.FromECDSAPub(&pk.PublicKey)
	did = "did:ethurl:123456"
	badSignature := "thisisabadsignature"
	verified, err = auth.VerifyEcdsaRequestSignature(string(pubKeyBys), badSignature,
		did, reqTs, auth.DefaultRequestGracePeriodSecs)
	if err == nil {
		t.Errorf("Should have returned error when verifying")
	}
	if verified {
		t.Errorf("Should have not verified this signature")
	}
}
