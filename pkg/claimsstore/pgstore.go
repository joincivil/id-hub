package claimsstore

import (
	"github.com/iden3/go-iden3-core/db"
	"github.com/jinzhu/gorm"
)

// PGStore is an implementation of the iden3 storage interface that uses postgres as its backend
type PGStore struct {
	NodePersister *NodePGPersister
	prefix        []byte
}

// NewPGStore returns a new postgress store
func NewPGStore(nodePersister *NodePGPersister) *PGStore {
	return &PGStore{
		NodePersister: nodePersister,
	}
}

// NewTx creates a new transaction
func (s *PGStore) NewTx() (db.Tx, error) {
	return &PGTX{s, make(kvMap)}, nil
}

// WithPrefix returns a new instance of the pgstore using the passed in prefix
func (s *PGStore) WithPrefix(prefix []byte) db.Storage {
	return &PGStore{NodePersister: s.NodePersister, prefix: concat(s.prefix, prefix)}
}

// Get gets the data from a node with the given key from the db
func (s *PGStore) Get(b []byte) ([]byte, error) {
	key := concat(s.prefix, b)

	value, err := s.NodePersister.Get(key)
	if gorm.IsRecordNotFoundError(err) {
		return nil, db.ErrNotFound
	}

	bvalue, err := value.ToDataBytes()

	return bvalue, err
}

// List lists all nodes for prefix/did
func (s *PGStore) List(limit int) ([]db.KV, error) {
	return s.NodePersister.GetAllForDID(s.prefix, limit)
}

// Close closes the db
func (s *PGStore) Close() {
	err := s.NodePersister.DB.Close()
	if err != nil {
		panic(err)
	}
}

// Info prints general information about all trees in the db
func (s *PGStore) Info() string {
	return s.NodePersister.Info()
}

// Iterate performs a function on all nodes in all trees
func (s *PGStore) Iterate(f func([]byte, []byte)) error {
	// WARNING iterate doesn't use the prefix

	// in the future probably want to do this in a way that doesn't load them all into memory
	kvs, err := s.NodePersister.GetAll()
	if err != nil {
		return err
	}

	for _, kv := range kvs {
		f(kv.K, kv.V)
	}
	return nil
}
