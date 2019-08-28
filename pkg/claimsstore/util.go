package claimsstore

import (
	"bytes"
)

// internal method from iden3 db that seemed necessary to implement the interface
func concat(vs ...[]byte) []byte {
	var b bytes.Buffer
	for _, v := range vs {
		b.Write(v)
	}
	return b.Bytes()
}
