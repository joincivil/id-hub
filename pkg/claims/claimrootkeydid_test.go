package claims_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	didlib "github.com/ockam-network/did"
	"github.com/stretchr/testify/assert"
)

func dataTestOutput(d *merkletree.Data) {
	s := bytes.NewBufferString("")
	fmt.Fprintf(s, "\t\t\"%v\"+\n", hex.EncodeToString(d[0][:]))
	fmt.Fprintf(s, "\t\t\"%v\"+\n", hex.EncodeToString(d[1][:]))
	fmt.Fprintf(s, "\t\t\"%v\"+\n", hex.EncodeToString(d[2][:]))
	fmt.Fprintf(s, "\t\t\"%v\",", hex.EncodeToString(d[3][:]))
	fmt.Println(s.String())
}

func TestClaimSetRootKey(t *testing.T) {
	dids := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	id, err := didlib.Parse(dids)
	assert.Nil(t, err)
	rootKey := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0c})
	c0, err := claims.NewClaimSetRootKeyDID(id, &rootKey)
	assert.Nil(t, err)
	c0.Version = 1
	c0.Era = 1
	e := c0.Entry()
	assert.Equal(t,
		"0x1c18c242b692cf2bf6a3989a5211ee28c0c129fc1f5f8d1c55f6b55b5c7ef554",
		e.HIndex().Hex())
	assert.Equal(t,
		"0x23af6c51c0ffe40d81508bf39e0360f884c9a1766895a8897a5e78da7bb611fa",
		e.HValue().Hex())
	dataTestOutput(&e.Data)
	assert.Equal(t, ""+
		"0000000000000000000000000000000000000000000000000000000000000000"+
		"0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0c"+
		"2036657468757269000000000000000086ce6c7127e64e0d83ddb60fe4d7785c"+
		"000000000000000000000000000000000000000100000001000000000000000a",
		e.Data.String())
	c1 := claims.NewClaimSetRootKeyDIDFromEntry(e)
	c2, err := claims.NewClaimFromEntry(e)
	assert.Nil(t, err)
	assert.Equal(t, c0, c1)
	assert.Equal(t, c0, c2)
	recoveredDID, err := claimsstore.BinaryToDID(c1.DID[:])
	assert.Nil(t, err)
	assert.Equal(t, recoveredDID.String(), dids)
}
