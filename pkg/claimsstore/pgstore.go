package claimsstore

import "github.com/iden3/go-iden3-core/db"

type PGStore struct {
}

func (s *PGStore) NewTx() (db.Tx, error) {

}
func (s *PGStore) WithPrefix(prefix []byte) db.Storage {

}
func (s *PGStore) Get(b []byte) ([]byte, error) {

}

func (s *PGStore) List(i int) ([]db.KV, error) {

}
func (s *PGStore) Close() {

}
func (s *PGStore) Info() string {

}
func (s *PGStore) Iterate(f func([]byte, []byte)) error {

}
