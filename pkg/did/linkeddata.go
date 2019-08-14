package did

import "time"

// Put all linked data code here for extraction later. Can add signing/verification
// code here as well.

// LinkedDataProof defines a linked data proof object
// Spec https://w3c-dvcg.github.io/ld-proofs/#linked-data-proof-overview
type LinkedDataProof struct {
	Type       string    `json:"type"`
	Creator    string    `json:"creator"`
	Created    time.Time `json:"created"`
	ProofValue string    `json:"proofValue"`
	Domain     *string   `json:"domain,omitempty"`
	Nonce      *string   `json:"nonce,omitempty"`
}
