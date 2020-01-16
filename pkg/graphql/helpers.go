package graphql

import (
	"encoding/json"
	"time"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/utils"
	"github.com/pkg/errors"
)

const (
	timeFormat = "2006-01-02T15:04:05Z"
)

// InputClaimToContentCredential is convenience function to convert input claim to
// content credential
func InputClaimToContentCredential(in *ClaimSaveRequestInput) (*claimtypes.ContentCredential, error) {
	var err error

	// if json blob was passed in, unmarshal and return
	if in.ClaimJSON != nil && *in.ClaimJSON != "" {
		cc := &claimtypes.ContentCredential{}
		err = json.Unmarshal([]byte(*in.ClaimJSON), cc)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal to content credential")
		}
		return cc, nil
	}

	if in.Claim == nil {
		return nil, errors.New("no claim passed in input")
	}

	cc := &claimtypes.ContentCredential{
		Context: in.Claim.Context,
		Issuer:  in.Claim.Issuer,
		Holder:  in.Claim.Holder,
	}

	cc.Type = ConvertCredentialTypes(in.Claim.Type)
	cc.CredentialSchema = ConvertCredentialSchema(*in.Claim.CredentialSchema)

	cc.CredentialSubject, err = ConvertCredentialSubject(in.Claim.CredentialSubject)
	if err != nil {
		return nil, err
	}

	// XXX(WF): This works assuming that whats sent in to the id hub is just one signed claim
	// this will need to become more complex in the future
	proof, err := ConvertInputProof(in.Claim.Proof[0])
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert input proof")
	}
	cc.Proof = []interface{}{*proof}

	// issuance date
	ts, err := time.Parse(timeFormat, in.Claim.IssuanceDate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the claim issuance date")
	}
	cc.IssuanceDate = ts

	return cc, nil
}

// ConvertCredentialTypes converts strings to credential types
func ConvertCredentialTypes(in []string) []claimtypes.CredentialType {
	types := make([]claimtypes.CredentialType, len(in))
	for ind, val := range in {
		types[ind] = claimtypes.CredentialType(val)
	}
	return types
}

// ConvertCredentialSchema converts input cred schema to core credential schema
func ConvertCredentialSchema(in ClaimCredentialSchemaInput) claimtypes.CredentialSchema {
	return claimtypes.CredentialSchema{
		ID:   in.ID,
		Type: in.Type,
	}
}

// ConvertCredentialSubject converts subject input to core credential subject
func ConvertCredentialSubject(in *ClaimCredentialSubjectInput) (
	claimtypes.ContentCredentialSubject, error) {
	if in == nil {
		return claimtypes.ContentCredentialSubject{},
			errors.New("no credential subject found")
	}

	md, err := ConvertArticleMetadata(in.Metadata)
	if err != nil {
		return claimtypes.ContentCredentialSubject{}, err
	}

	return claimtypes.ContentCredentialSubject{
		ID:       in.ID,
		Metadata: *md,
	}, nil
}

// ConvertArticleMetadata converts incoming article metadata and converts it to
// core article metadata object
func ConvertArticleMetadata(in *ArticleMetadataInput) (*article.Metadata, error) {
	md := &article.Metadata{
		Title:               *in.Title,
		RevisionContentHash: *in.RevisionContentHash,
		RevisionContentURL:  *in.RevisionContentURL,
		CanonicalURL:        *in.CanonicalURL,
		Slug:                *in.Slug,
		Description:         *in.Description,
		PrimaryTag:          *in.PrimaryTag,
		Opinion:             *in.Opinion,
		CivilSchemaVersion:  *in.CivilSchemaVersion,
	}

	contribs := make([]article.Contributor, len(in.Contributors))
	for ind, v := range in.Contributors {
		contribs[ind] = article.Contributor{
			Role: *v.Role,
			Name: *v.Name,
		}
	}
	md.Contributors = contribs

	images := make([]article.Image, len(in.Images))
	for ind, v := range in.Images {
		images[ind] = article.Image{
			URL:  *v.URL,
			Hash: *v.Hash,
			H:    *v.H,
			W:    *v.W,
		}
	}
	md.Images = images

	ts, err := time.Parse(timeFormat, *in.RevisionDate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the revision date")
	}
	md.RevisionDate = ts

	ts, err = time.Parse(timeFormat, *in.OriginalPublishDate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the original pub date")
	}
	md.OriginalPublishDate = ts

	return md, nil
}

// ConvertInputProof validates and converts linked data proof input to
// a core linked data proof.
func ConvertInputProof(in *LinkedDataProofInput) (*linkeddata.Proof, error) {
	if in == nil {
		return nil, nil
	}

	ldp := &linkeddata.Proof{}

	ldp.Type = utils.StrOrEmptyStr(in.Type)
	ldp.Creator = utils.StrOrEmptyStr(in.Creator)
	ldp.ProofValue = utils.StrOrEmptyStr(in.ProofValue)
	ldp.Domain = in.Domain
	ldp.Nonce = in.Nonce

	if in.Created != nil {
		ldp.Created = *in.Created
	}

	return ldp, nil
}
