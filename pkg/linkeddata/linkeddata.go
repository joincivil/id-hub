package linkeddata

import "time"

// Put all linked data code here for extraction later. Can add signing/verification
// code here as well.

// SuiteType defines a type for LD crypto suite type
type SuiteType string

const (
	// SuiteTypeRsaSignature defines LD crypto suite type for RSA signatures
	SuiteTypeRsaSignature SuiteType = "RsaSignature2018"
	// SuiteTypeRsaVerification defines LD crypto suite type for RSA verifications
	SuiteTypeRsaVerification SuiteType = "RsaVerificationKey2018"
	// SuiteTypeSecp256k1Signature defines LD crypto suite type for Secp256k signatures
	SuiteTypeSecp256k1Signature SuiteType = "EcdsaSecp256k1Signature2019"
	// SuiteTypeSecp256k1Verification defines LD crypto suite type for Secp256k verifications
	SuiteTypeSecp256k1Verification SuiteType = "EcdsaSecp256k1VerificationKey2019"
	// SuiteTypeEd25519Signature defines LD crypto suite type for Ed25519 signatures
	SuiteTypeEd25519Signature SuiteType = "Ed25519Signature2018"
	// SuiteTypeEd25519Verification defines LD crypto suite type for Ed25519 verifications
	SuiteTypeEd25519Verification SuiteType = "Ed25519VerificationKey2018"
	// SuiteTypeKoblitzSignature defines a LD crypto suite type for Koblitz signatures
	SuiteTypeKoblitzSignature SuiteType = "EcdsaKoblitzSignature2016"
)

// IsEcdsaKeySuiteType returns true if key cryptographic suite type is of elliptic curve,
// namely secp251k1
func IsEcdsaKeySuiteType(keytype SuiteType) bool {
	switch keytype {
	case SuiteTypeSecp256k1Signature:
		return true
	case SuiteTypeSecp256k1Verification:
		return true
	}
	return false
}

// Proof defines a linked data proof object
// Spec https://w3c-dvcg.github.io/ld-proofs/#linked-data-proof-overview
type Proof struct {
	Type       string    `json:"type"`
	Creator    string    `json:"creator"`
	Created    time.Time `json:"created"`
	ProofValue string    `json:"proofValue,omitempty"`
	Domain     *string   `json:"domain,omitempty"`
	Nonce      *string   `json:"nonce,omitempty"`
}
