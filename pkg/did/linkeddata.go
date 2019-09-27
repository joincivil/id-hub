package did

import "time"

// Put all linked data code here for extraction later. Can add signing/verification
// code here as well.

// LDSuiteType defines a type for LD crypto suite type
type LDSuiteType string

const (
	// LDSuiteTypeRsaSignature defines LD crypto suite type for RSA signatures
	LDSuiteTypeRsaSignature LDSuiteType = "RsaSignature2018"
	// LDSuiteTypeRsaVerification defines LD crypto suite type for RSA verifications
	LDSuiteTypeRsaVerification LDSuiteType = "RsaVerificationKey2018"
	// LDSuiteTypeSecp256k1Signature defines LD crypto suite type for Secp256k signatures
	LDSuiteTypeSecp256k1Signature LDSuiteType = "EcdsaSecp256k1Signature2019"
	// LDSuiteTypeSecp256k1Verification defines LD crypto suite type for Secp256k verifications
	LDSuiteTypeSecp256k1Verification LDSuiteType = "EcdsaSecp256k1VerificationKey2019"
	// LDSuiteTypeEd25519Signature defines LD crypto suite type for Ed25519 signatures
	LDSuiteTypeEd25519Signature LDSuiteType = "Ed25519Signature2018"
	// LDSuiteTypeEd25519Verification defines LD crypto suite type for Ed25519 verifications
	LDSuiteTypeEd25519Verification LDSuiteType = "Ed25519VerificationKey2018"
	// LDSuiteTypeKoblitzSignature defines a LD crypto suite type for Koblitz signatures
	LDSuiteTypeKoblitzSignature LDSuiteType = "EcdsaKoblitzSignature2016"
)

// IsEcdsaKeySuiteType returns true if key cryptographic suite type is of elliptic curve,
// namely secp251k1
func IsEcdsaKeySuiteType(keytype LDSuiteType) bool {
	switch keytype {
	case LDSuiteTypeSecp256k1Signature:
		return true
	case LDSuiteTypeSecp256k1Verification:
		return true
	}
	return false
}

// LinkedDataProof defines a linked data proof object
// Spec https://w3c-dvcg.github.io/ld-proofs/#linked-data-proof-overview
type LinkedDataProof struct {
	Type       string    `json:"type"`
	Creator    string    `json:"creator"`
	Created    time.Time `json:"created"`
	ProofValue string    `json:"proofValue,omitempty"`
	Domain     *string   `json:"domain,omitempty"`
	Nonce      *string   `json:"nonce,omitempty"`
}
