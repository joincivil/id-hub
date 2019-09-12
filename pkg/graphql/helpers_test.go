package graphql_test

import (
	"testing"
	"time"

	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/utils"
)

func TestValidateConvertInputPublicKeys(t *testing.T) {
	inputPk1 := &graphql.DidDocPublicKeyInput{
		ID:           utils.StrToPtr("did:ethuri:123456"),
		Controller:   utils.StrToPtr("did:ethuri:123456"),
		Type:         utils.StrToPtr("EcdsaSecp256k1VerificationKey2019"),
		PublicKeyHex: utils.StrToPtr("04debef3fcbef3f5659f9169bad80044b287139a401b5da2979e50b032560ed33927eab43338e9991f31185b3152735e98e0471b76f18897d764b4e4f8a7e8f61b"),
	}
	inputPks := []*graphql.DidDocPublicKeyInput{inputPk1}

	pks, pkMap, err := graphql.ValidateConvertInputPublicKeys(inputPks)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
	if len(pks) != 1 {
		t.Errorf("Should have gotten 1 pks")
	}
	if len(pkMap) != 1 {
		t.Errorf("Should have gotten 1 items in the pk map: len: %v", len(pkMap))
	}
}

func TestValidateConvertInputAuthentications(t *testing.T) {
	inputPk1 := &graphql.DidDocPublicKeyInput{
		ID:           utils.StrToPtr("did:ethuri:123456"),
		Controller:   utils.StrToPtr("did:ethuri:123456"),
		Type:         utils.StrToPtr("EcdsaSecp256k1VerificationKey2019"),
		PublicKeyHex: utils.StrToPtr("04debef3fcbef3f5659f9169bad80044b287139a401b5da2979e50b032560ed33927eab43338e9991f31185b3152735e98e0471b76f18897d764b4e4f8a7e8f61b"),
	}
	idOnly := false
	inputAuth1 := &graphql.DidDocAuthenticationInput{
		PublicKey: inputPk1,
		IDOnly:    &idOnly,
	}
	inputAuths := []*graphql.DidDocAuthenticationInput{inputAuth1}

	auths, err := graphql.ValidateConvertInputAuthentications(inputAuths, map[string]int{})
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
	if len(auths) != 1 {
		t.Errorf("Should have gotten 1 auth")
	}
}

func TestValidateConvertInputServices(t *testing.T) {
	inputSrv1 := &graphql.DidDocServiceInput{
		ID:              utils.StrToPtr("did:ethuri:123456#vcr"),
		Type:            utils.StrToPtr("CredentialRepositoryService"),
		Description:     utils.StrToPtr("This is a description"),
		ServiceEndpoint: &utils.AnyValue{Value: "https://repository.example.com/service/8377464"},
	}
	inputSrvs := []*graphql.DidDocServiceInput{inputSrv1}

	srvs, err := graphql.ValidateConvertInputServices(inputSrvs)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
	if len(srvs) != 1 {
		t.Errorf("Should have gotten 1 srv")
	}
}

func TestValidateConvertInputProof(t *testing.T) {
	ts := time.Now()
	inputProof := &graphql.LinkedDataProofInput{
		Type:       utils.StrToPtr("LinkedDataSignature2015"),
		Creator:    utils.StrToPtr("did:ethuri:123456"),
		Created:    &ts,
		ProofValue: utils.StrToPtr("thisisasignature value"),
	}

	_, err := graphql.ValidateConvertInputProof(inputProof)
	if err != nil {
		t.Errorf("Should not have gotten error: err: %v", err)
	}
}
