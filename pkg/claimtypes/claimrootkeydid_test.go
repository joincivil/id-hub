package claimtypes_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/joincivil/id-hub/pkg/claimtypes"
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
	c0, err := claimtypes.NewClaimSetRootKeyDID(id, &rootKey)
	assert.Nil(t, err)
	c0.Version = 1
	c0.Era = 1
	e := c0.Entry()
	assert.Equal(t,
		"0x2e50a8b631ab4a3c1563155e7e007ccbc5e250b24b3391af2aecf3fea8b63186",
		e.HIndex().Hex())
	assert.Equal(t,
		"0x23af6c51c0ffe40d81508bf39e0360f884c9a1766895a8897a5e78da7bb611fa",
		e.HValue().Hex())
	dataTestOutput(&e.Data)
	assert.Equal(t, ""+
		"0000000000000000000000000000000000000000000000000000000000000000"+
		"0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0c"+
		"22fec3525a9169b13d4452bf5c528ccf5ba2ab6cc159bb74e28d4b29b682d641"+
		"000000000000000000000000000000000000000100000001000000000000000a",
		e.Data.String())
	c1 := claimtypes.NewClaimSetRootKeyDIDFromEntry(e)
	c2, err := claimtypes.NewClaimFromEntry(e)
	assert.Nil(t, err)
	assert.Equal(t, c0, c1)
	assert.Equal(t, c0, c2)
}
