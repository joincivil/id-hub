package graphql_test

import (
	"strings"
	"testing"

	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/graphql"
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
