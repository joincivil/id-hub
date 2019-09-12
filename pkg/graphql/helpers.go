package graphql

import (
	"fmt"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/utils"
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
		Type:               did.LDSuiteType(*in.Type),
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

// ValidateConvertInputPublicKeys validates and converts public key input to core
// did.DocPublicKey values. Also returns a map of public key ids for lookup
func ValidateConvertInputPublicKeys(in []*DidDocPublicKeyInput) ([]did.DocPublicKey,
	map[string]int, error) {
	if in == nil {
		return nil, nil, nil
	}

	var pk *did.DocPublicKey
	pks := make([]did.DocPublicKey, len(in))
	pkMap := map[string]int{}

	for ind, inPk := range in {
		pk = InputPkToDocPublicKey(inPk)
		if !did.ValidDocPublicKey(pk) {
			return nil, nil, errors.New("pk is invalid")
		}

		if pk.ID != nil {
			pkMap[pkMapKey(pk)] = 1
		}

		pks[ind] = *pk
	}

	return pks, pkMap, nil
}

// ValidateConvertInputAuthentications validates and converts auth key input to core
// did.DocAuthenticationWrapper values. Take map of pk to verify they exist for use
// with auth.
func ValidateConvertInputAuthentications(in []*DidDocAuthenticationInput,
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
		} else if !did.ValidDocPublicKey(&auth.DocPublicKey) {
			return nil, errors.Errorf("auth pk is invalid: %v", authDid)
		}
		auths[ind] = *auth
	}

	return auths, nil
}

// ValidateConvertInputServices validates and converts input services to core did.DocService
// objects.
func ValidateConvertInputServices(in []*DidDocServiceInput) ([]did.DocService, error) {
	if in == nil {
		return nil, nil
	}

	var srv *did.DocService
	srvs := make([]did.DocService, len(in))

	// TODO(PN): validation

	for ind, inAuth := range in {
		srv = InputServiceToDocService(inAuth)
		srvs[ind] = *srv
	}

	return srvs, nil
}

// ValidateConvertInputProof validates and converts linked data proof input to
// a core linked data proof.
func ValidateConvertInputProof(in *LinkedDataProofInput) (*did.LinkedDataProof, error) {
	if in == nil {
		return nil, nil
	}

	ldp := &did.LinkedDataProof{}

	// TODO(PN): validation

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
	return fmt.Sprintf("%v-%v", pk.ID.String(), pk.Type)
}
