package graphql

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	didlib "github.com/ockam-network/did"

	"github.com/joincivil/go-common/pkg/article"
	cstore "github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/utils"
)

const (
	timeFormat = "2006-01-02T15:04:05Z"
)

// InputPkToDocPublicKey is convenience function to convert input public key values
// into a did.DocPublicKey object
func InputPkToDocPublicKey(in *DidDocPublicKeyInput) *did.DocPublicKey {
	var err error
	var theDid *didlib.DID
	var controller *didlib.DID

	if in.ID != nil {
		theDid, err = didlib.Parse(*in.ID)
		if err != nil {
			log.Errorf("Error parsing DID: err: %v", err)
		}
	}

	if in.Controller != nil {
		controller, err = didlib.Parse(*in.Controller)
		if err != nil {
			log.Errorf("Error parsing controller did: err: %v", err)
		}
	}

	return &did.DocPublicKey{
		ID:                 theDid,
		Type:               linkeddata.SuiteType(*in.Type),
		Controller:         controller,
		PublicKeyPem:       in.PublicKeyPem,
		PublicKeyJwk:       in.PublicKeyJwk,
		PublicKeyHex:       in.PublicKeyHex,
		PublicKeyBase64:    in.PublicKeyBase64,
		PublicKeyBase58:    in.PublicKeyBase58,
		PublicKeyMultibase: in.PublicKeyMultibase,
		EthereumAddress:    in.EthereumAddress,
	}
}

// InputAuthToDocAuthentication is convenience function to convert input auth key values
// into a did.DocAuthenticationWrapper object
func InputAuthToDocAuthentication(in *DidDocAuthenticationInput) *did.DocAuthenicationWrapper {
	if in == nil {
		return nil
	}

	idOnly := false
	if in.IDOnly != nil {
		idOnly = *in.IDOnly
	}
	return &did.DocAuthenicationWrapper{
		DocPublicKey: *InputPkToDocPublicKey(in.PublicKey),
		IDOnly:       idOnly,
	}
}

// InputServiceToDocService is convenience function to convert input service values
// into a did.DocService object
func InputServiceToDocService(in *DidDocServiceInput) *did.DocService {
	if in == nil {
		return nil
	}

	srv := &did.DocService{
		Type:        utils.StrOrEmptyStr(in.Type),
		Description: utils.StrOrEmptyStr(in.Description),
		PublicKey:   utils.StrOrEmptyStr(in.PublicKey),
	}

	if in.ID != nil {
		d, err := didlib.Parse(*in.ID)
		if err != nil {
			log.Errorf("Error parsing did in service: err: %v", err)
		}
		srv.ID = *d
	}

	if in.ServiceEndpoint != nil {
		srv.ServiceEndpoint = in.ServiceEndpoint.Value
		err := srv.PopulateServiceEndpointVals()
		if err != nil {
			log.Errorf("Error populating service endpoint values: err: %v", err)
		}
	}

	return srv
}

// InputClaimToContentCredential is convenience function to convert input claim to
// content credential
func InputClaimToContentCredential(in *ClaimSaveRequestInput) (*cstore.ContentCredential, error) {
	var err error

	// if json blob was passed in, unmarshal and return
	if in.ClaimJSON != nil && *in.ClaimJSON != "" {
		cc := &cstore.ContentCredential{}
		err = json.Unmarshal([]byte(*in.ClaimJSON), cc)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal to content credential")
		}
		return cc, nil
	}

	if in.Claim == nil {
		return nil, errors.New("no claim passed in input")
	}

	cc := &cstore.ContentCredential{
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

	proof, err := ConvertInputProof(in.Claim.Proof)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert input proof")
	}
	cc.Proof = *proof

	// issuance date
	ts, err := time.Parse(timeFormat, in.Claim.IssuanceDate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the claim issuance date")
	}
	cc.IssuanceDate = ts

	return cc, nil
}

// ConvertCredentialTypes converts strings to credential types
func ConvertCredentialTypes(in []string) []cstore.CredentialType {
	types := make([]cstore.CredentialType, len(in))
	for ind, val := range in {
		types[ind] = cstore.CredentialType(val)
	}
	return types
}

// ConvertCredentialSchema converts input cred schema to core credential schema
func ConvertCredentialSchema(in ClaimCredentialSchemaInput) cstore.CredentialSchema {
	return cstore.CredentialSchema{
		ID:   in.ID,
		Type: in.Type,
	}
}

// ConvertCredentialSubject converts subject input to core credential subject
func ConvertCredentialSubject(in *ClaimCredentialSubjectInput) (
	cstore.ContentCredentialSubject, error) {
	if in == nil {
		return cstore.ContentCredentialSubject{},
			errors.New("no credential subject found")
	}

	md, err := ConvertArticleMetadata(in.Metadata)
	if err != nil {
		return cstore.ContentCredentialSubject{}, err
	}

	return cstore.ContentCredentialSubject{
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

// ConvertInputPublicKeys validates and converts public key input to core
// did.DocPublicKey values. Also returns a map of public key ids for lookup
func ConvertInputPublicKeys(in []*DidDocPublicKeyInput) ([]did.DocPublicKey,
	map[string]int, error) {
	if in == nil {
		return nil, nil, nil
	}

	var pk *did.DocPublicKey
	pks := make([]did.DocPublicKey, len(in))
	pkMap := map[string]int{}

	for ind, inPk := range in {
		pk = InputPkToDocPublicKey(inPk)
		if pk.ID != nil {
			pkMap[pkMapKey(pk)] = 1
		}

		pks[ind] = *pk
	}

	return pks, pkMap, nil
}

// ConvertInputAuthentications validates and converts auth key input to core
// did.DocAuthenticationWrapper values. Take map of pk to verify they exist for use
// with auth.
func ConvertInputAuthentications(in []*DidDocAuthenticationInput,
	pkMap map[string]int) ([]did.DocAuthenicationWrapper, error) {
	if in == nil {
		return nil, nil
	}

	var auth *did.DocAuthenicationWrapper
	auths := make([]did.DocAuthenicationWrapper, len(in))
	var authDid string
	for ind, inAuth := range in {
		auth = InputAuthToDocAuthentication(inAuth)
		authDid = auth.DocPublicKey.ID.String()
		if auth.IDOnly {
			if pkMap != nil {
				_, ok := pkMap[pkMapKey(&auth.DocPublicKey)]
				if !ok {
					return nil, errors.Errorf("auth pk is not in public keys: %v", authDid)
				}
			}
		}
		auths[ind] = *auth
	}

	return auths, nil
}

// ConvertInputServices validates and converts input services to core did.DocService
// objects.
func ConvertInputServices(in []*DidDocServiceInput) ([]did.DocService, error) {
	if in == nil {
		return nil, nil
	}

	var srv *did.DocService
	srvs := make([]did.DocService, len(in))

	for ind, inAuth := range in {
		srv = InputServiceToDocService(inAuth)
		srvs[ind] = *srv
	}

	return srvs, nil
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

func pkMapKey(pk *did.DocPublicKey) string {
	id := "new"
	if pk.ID != nil && pk.ID.String() != "" {
		id = pk.ID.String()
	}
	return fmt.Sprintf("%v-%v", id, pk.Type)
}
