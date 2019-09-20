package did_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
)

func check(persister did.Persister) {
}

func interfaceCheck() {
	check(&did.InMemoryPersister{})
}

func TestInterface(t *testing.T) {
	interfaceCheck()
}
