package graphql_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	ctime "github.com/joincivil/go-common/pkg/time"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/didjwt"
	didlib "github.com/ockam-network/did"
)

func TestAddEdge(t *testing.T) {
	resolver, rootService, ethURI, err := setupResolver(t)
	if err != nil {
		t.Errorf("couldn't create resolver")
	}
	claimerDid, err := didlib.Parse("did:ethuri:cc4ef0ec-bd37-46e6-8419-3164c325205f")
	if err != nil {
		t.Errorf("couldn't parse did: %v", err)
	}

	if err := createDID(ethURI, claimerDid); err != nil {
		t.Errorf("couldn't add the did: %v", err)
	}

	claims := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer: claimerDid.String(),
		},
	}

	pk, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Errorf("couldn't create private key: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenS, err := token.SignedString(pk)
	if err != nil {
		t.Errorf("unable to create jwt string: %v", err)
	}

	reqTs := ctime.CurrentEpochSecsInInt()
	queries := resolver.Query()
	mutations := resolver.Mutation()

	_, err = mutations.AddEdge(context.Background(), &tokenS)
	if err == nil {
		t.Errorf("should have errored on auth")
	}

	signature, err := auth.SignEcdsaRequestMessage(pk, claimerDid.String(), reqTs)
	if err != nil {
		t.Errorf("couldn't create auth signature: %v", err)
	}

	c := context.Background()
	c = context.WithValue(c, auth.ReqTsCtxKey, strconv.Itoa(reqTs))
	c = context.WithValue(c, auth.DidCtxKey, claimerDid.String())
	c = context.WithValue(c, auth.SignatureCtxKey, signature)

	edge, err := mutations.AddEdge(c, &tokenS)
	if err != nil {
		t.Errorf("unexpected err save the claim: %v", err)
	}

	if edge.JWT != tokenS {
		t.Errorf("bad response")
	}

	// commit the root
	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing root: %v", err)
	}

	edges, err := queries.FindEdges(c, claimerDid.String())
	if err != nil {
		t.Errorf("error finding edges for did: %v", err)
	}

	if len(edges) != 1 {
		t.Errorf("wrong number of edges expected: 1 got: %v", len(edges))
	}
}

func TestProof(t *testing.T) {

}
