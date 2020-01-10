package ethuri_test

import (
	"testing"

	"github.com/joincivil/id-hub/pkg/did/ethuri"
)

func check(persister ethuri.Persister) {
}

func interfaceCheck() {
	check(&ethuri.InMemoryPersister{})
}

func TestInterface(t *testing.T) {
	interfaceCheck()
}
