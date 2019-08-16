package claimsstore

import "github.com/iden3/go-iden3-core/db"

type PGTX struct {
}

func (t *PGTX) Get(b []byte) ([]byte, error) {

}
func (t *PGTX) Put(k, v []byte) {

}
func (t *PGTX) Add(db.Tx) {

}
func (t *PGTX) Commit() error {

}
func (t *PGTX) Close() {

}
