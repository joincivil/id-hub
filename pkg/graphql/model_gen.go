// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	"time"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/utils"
)

type Proof interface {
	IsProof()
}

type ArticleMetadataContributorInput struct {
	Role *string `json:"role"`
	Name *string `json:"name"`
}

type ArticleMetadataImageInput struct {
	URL  *string `json:"url"`
	Hash *string `json:"hash"`
	H    *int    `json:"h"`
	W    *int    `json:"w"`
}

type ArticleMetadataInput struct {
	Title               *string                            `json:"title"`
	RevisionContentHash *string                            `json:"revisionContentHash"`
	RevisionContentURL  *string                            `json:"revisionContentURL"`
	CanonicalURL        *string                            `json:"canonicalURL"`
	Slug                *string                            `json:"slug"`
	Description         *string                            `json:"description"`
	Contributors        []*ArticleMetadataContributorInput `json:"contributors"`
	Images              []*ArticleMetadataImageInput       `json:"images"`
	Tags                []*string                          `json:"tags"`
	PrimaryTag          *string                            `json:"primaryTag"`
	RevisionDate        *string                            `json:"revisionDate"`
	OriginalPublishDate *string                            `json:"originalPublishDate"`
	Opinion             *bool                              `json:"opinion"`
	CivilSchemaVersion  *string                            `json:"civilSchemaVersion"`
}

type ClaimCredentialSchemaInput struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ClaimCredentialSubjectInput struct {
	ID       string                `json:"id"`
	Metadata *ArticleMetadataInput `json:"metadata"`
}

type ClaimGetRequestInput struct {
	Did string `json:"did"`
}

type ClaimInput struct {
	Context           []string                     `json:"context"`
	Type              []string                     `json:"type"`
	CredentialSubject *ClaimCredentialSubjectInput `json:"credentialSubject"`
	Issuer            string                       `json:"issuer"`
	Holder            string                       `json:"holder"`
	CredentialSchema  *ClaimCredentialSchemaInput  `json:"credentialSchema"`
	IssuanceDate      string                       `json:"issuanceDate"`
	Proof             []*LinkedDataProofInput      `json:"proof"`
}

type ClaimProofRequestInput struct {
	Claim     *ClaimInput `json:"claim"`
	ClaimJSON *string     `json:"claimJson"`
	Did       string      `json:"did"`
}

type ClaimProofResponse struct {
	Claim    *claimtypes.ContentCredential `json:"claim"`
	ClaimRaw string                        `json:"claimRaw"`
}

type ClaimRegisteredProof struct {
	Type                   string `json:"type"`
	Did                    string `json:"did"`
	ExistsInDIDMTProof     string `json:"existsInDIDMTProof"`
	NotRevokedInDIDMTProof string `json:"notRevokedInDIDMTProof"`
	DidMTRootExistsProof   string `json:"didMTRootExistsProof"`
	DidRootExistsVersion   int    `json:"didRootExistsVersion"`
	Root                   string `json:"root"`
	DidMTRoot              string `json:"didMTRoot"`
}

func (ClaimRegisteredProof) IsProof() {}

type ClaimSaveRequestInput struct {
	Claim     *ClaimInput `json:"claim"`
	ClaimJSON *string     `json:"claimJson"`
}

type ContentClaimCredentialSubject struct {
	ID       string            `json:"id"`
	Metadata *article.Metadata `json:"metadata"`
}

type DidDocAuthenticationInput struct {
	PublicKey *DidDocPublicKeyInput `json:"publicKey"`
	IDOnly    *bool                 `json:"idOnly"`
}

type DidDocPublicKeyInput struct {
	ID                 *string `json:"id"`
	Type               *string `json:"type"`
	Controller         *string `json:"controller"`
	PublicKeyPem       *string `json:"publicKeyPem"`
	PublicKeyJwk       *string `json:"publicKeyJwk"`
	PublicKeyHex       *string `json:"publicKeyHex"`
	PublicKeyBase64    *string `json:"publicKeyBase64"`
	PublicKeyBase58    *string `json:"publicKeyBase58"`
	PublicKeyMultibase *string `json:"publicKeyMultibase"`
	EthereumAddress    *string `json:"ethereumAddress"`
}

type DidDocServiceInput struct {
	ID              *string         `json:"id"`
	Type            *string         `json:"type"`
	Description     *string         `json:"description"`
	PublicKey       *string         `json:"publicKey"`
	ServiceEndpoint *utils.AnyValue `json:"serviceEndpoint"`
}

type DidGetRequestInput struct {
	Did *string `json:"did"`
}

type DidSaveRequestInput struct {
	Did             *string                      `json:"did"`
	PublicKeys      []*DidDocPublicKeyInput      `json:"publicKeys"`
	Authentications []*DidDocAuthenticationInput `json:"authentications"`
	Services        []*DidDocServiceInput        `json:"services"`
	Proof           *LinkedDataProofInput        `json:"proof"`
}

type LinkedDataProofInput struct {
	Type       *string    `json:"type"`
	Creator    *string    `json:"creator"`
	Created    *time.Time `json:"created"`
	ProofValue *string    `json:"proofValue"`
	Domain     *string    `json:"domain"`
	Nonce      *string    `json:"nonce"`
}

type RootOnBlockChainProof struct {
	Type             string `json:"type"`
	BlockNumber      string `json:"blockNumber"`
	Root             string `json:"root"`
	ContractAddress  string `json:"contractAddress"`
	CommitterAddress string `json:"committerAddress"`
	TxHash           string `json:"txHash"`
}

func (RootOnBlockChainProof) IsProof() {}
