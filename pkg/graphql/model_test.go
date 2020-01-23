package graphql_test

import (
	"strings"
	"testing"
	"time"

	didlib "github.com/ockam-network/did"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/utils"
)

func TestDidGetResponse(t *testing.T) {
	d, _ := didlib.Parse("did:ethuri:123456")

	doc := &did.Document{
		Context: did.DefaultDIDContextV1,
		ID:      *d,
	}
	resp := &graphql.DidGetResponse{Doc: doc}

	raw := resp.DocRaw()
	if raw == nil {
		t.Errorf("Should not have returned nil for raw")
	}
	if !strings.Contains(*raw, "\"id\"") {
		t.Errorf("Should have had an id field id")
	}
	if !strings.Contains(*raw, "\"@context\"") {
		t.Errorf("Should have a @context field")
	}
}

func TestDidSaveResponse(t *testing.T) {
	d, _ := didlib.Parse("did:ethuri:123456")

	doc := &did.Document{
		Context: did.DefaultDIDContextV1,
		ID:      *d,
	}
	resp := &graphql.DidSaveResponse{Doc: doc}

	raw := resp.DocRaw()
	if raw == nil {
		t.Errorf("Should not have returned nil for raw")
	}
	if !strings.Contains(*raw, "\"id\"") {
		t.Errorf("Should have had an id field id")
	}
	if !strings.Contains(*raw, "\"@context\"") {
		t.Errorf("Should have a @context field")
	}
}

func TestClaimGetResponse(t *testing.T) {
	now := time.Now()
	claims := []*claimtypes.ContentCredential{
		{
			Context: []string{
				"https://www.w3.org/2018/credentials/v1",
				"https://id.civil.co/credentials/contentcredential/v1",
			},
			Type: []claimtypes.CredentialType{
				claimtypes.VerifiableCredentialType,
				claimtypes.ContentCredentialType,
			},
			CredentialSubject: claimtypes.ContentCredentialSubject{
				ID: "did:web:civil.co",
				Metadata: article.Metadata{
					Title:               "Test Title",
					RevisionContentHash: "",
					RevisionContentURL:  "",
					CanonicalURL:        "",
					Slug:                "",
					Description:         "",
					PrimaryTag:          "",
					Opinion:             false,
					CivilSchemaVersion:  "",
					RevisionDate:        time.Now(),
					OriginalPublishDate: time.Now(),
					Contributors: []article.Contributor{
						{
							Name: "Eric Martin",
							Role: "author",
						},
					},
					Images: []article.Image{
						{
							URL:  "http://civil.co/img.jpg",
							Hash: "iamahash",
							W:    10,
							H:    10,
						},
					},
				},
			},
			CredentialSchema: claimtypes.CredentialSchema{
				ID:   "https://id.civil.co/credentials/schemas/v1/metadata.json",
				Type: "JsonSchemaValidator2018",
			},
			Proof: []*graphql.LinkedDataProofInput{
				{
					Type:       utils.StrToPtr("EcdsaSecp256k1Signature2019"),
					Creator:    utils.StrToPtr("did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1"),
					Created:    &now,
					ProofValue: utils.StrToPtr("80ed91bd852ba71ef230b74acb66375fe1516c6e282cb202fe10dcf6c0cc14934c179b88a3da16e3737283cd597732936bc51631bdc2596edc2d82d65d9610f900"),
				},
			},
			Issuer:       "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1",
			IssuanceDate: time.Now(),
		},
	}
	resp := &graphql.ClaimGetResponse{Claims: claims}
	raw := resp.ClaimsRaw()
	for _, s := range raw {
		if !strings.Contains(s, "\"id\"") {
			t.Errorf("Should have had an id field id")
		}
		if !strings.Contains(s, "\"@context\"") {
			t.Errorf("Should have a @context field")
		}
	}
}

func TestClaimSaveResponse(t *testing.T) {
	now := time.Now()
	claim := &claimtypes.ContentCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://id.civil.co/credentials/contentcredential/v1",
		},
		Type: []claimtypes.CredentialType{
			claimtypes.VerifiableCredentialType,
			claimtypes.ContentCredentialType,
		},
		CredentialSubject: claimtypes.ContentCredentialSubject{
			ID: "did:web:civil.co",
			Metadata: article.Metadata{
				Title:               "Test Title",
				RevisionContentHash: "",
				RevisionContentURL:  "",
				CanonicalURL:        "",
				Slug:                "",
				Description:         "",
				PrimaryTag:          "",
				Opinion:             false,
				CivilSchemaVersion:  "",
				RevisionDate:        time.Now(),
				OriginalPublishDate: time.Now(),
				Contributors: []article.Contributor{
					{
						Name: "Eric Martin",
						Role: "author",
					},
				},
				Images: []article.Image{
					{
						URL:  "http://civil.co/img.jpg",
						Hash: "iamahash",
						W:    10,
						H:    10,
					},
				},
			},
		},
		CredentialSchema: claimtypes.CredentialSchema{
			ID:   "https://id.civil.co/credentials/schemas/v1/metadata.json",
			Type: "JsonSchemaValidator2018",
		},
		Proof: []*graphql.LinkedDataProofInput{
			{
				Type:       utils.StrToPtr("EcdsaSecp256k1Signature2019"),
				Creator:    utils.StrToPtr("did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1"),
				Created:    &now,
				ProofValue: utils.StrToPtr("80ed91bd852ba71ef230b74acb66375fe1516c6e282cb202fe10dcf6c0cc14934c179b88a3da16e3737283cd597732936bc51631bdc2596edc2d82d65d9610f900"),
			},
		},
		Issuer:       "did:ethuri:c81773d7-fc03-44eb-b28e-6eff2c291485#keys-1",
		IssuanceDate: time.Now(),
	}

	resp := &graphql.ClaimSaveResponse{Claim: claim}
	raw := resp.ClaimRaw()
	if !strings.Contains(*raw, "\"id\"") {
		t.Errorf("Should have had an id field id")
	}
	if !strings.Contains(*raw, "\"@context\"") {
		t.Errorf("Should have a @context field")
	}
}
