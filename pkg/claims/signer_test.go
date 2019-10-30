package claims_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/claims"
	didlib "github.com/ockam-network/did"
)

func TestSignerSign(t *testing.T) {
	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDIDKey := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c#keys-1"
	userDID, err := didlib.Parse(userDIDs)
	if err != nil {
		t.Fatalf("error parsing did: %v", err)
	}
	key, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Fatalf("should be able to make a key")
	}
	ecdsaSigner := claims.NewECDSASigner(key)
	claim := makeContentCredential(userDID)
	err = ecdsaSigner.Sign(claim, userDIDKey)
	if err != nil {
		t.Errorf("should not have errored creating proof: %v", err)
	}
	canoncred, err := claims.CanonicalizeCredential(claim)
	if err != nil {
		t.Errorf("error canonicalizing the claim: %v", err)
	}
	sigbytes, err := hex.DecodeString(claim.Proof.ProofValue)
	if err != nil {
		t.Errorf("error decoding signature: %v", err)
	}
	pubBytes := crypto.FromECDSAPub(&key.PublicKey)

	recoveredPubkey, err := crypto.SigToPub(crypto.Keccak256(canoncred), sigbytes)
	if err != nil {
		t.Errorf("could not recover public key: %v", err)
	}
	recoveredBytes := crypto.FromECDSAPub(recoveredPubkey)

	if !bytes.Equal(recoveredBytes, pubBytes) {
		t.Errorf("could not verify the signature")
	}
}
