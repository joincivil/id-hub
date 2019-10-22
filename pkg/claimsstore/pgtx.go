package claimsstore

import (
	"crypto/sha256"

	"github.com/iden3/go-iden3-core/db"
	"github.com/jinzhu/gorm"
)

func newKvMap() *kvMap {
	return &kvMap{
		kv:    map[[sha256.Size]byte]db.KV{},
		order: [][sha256.Size]byte{},
	}
}

type kvMap struct {
	kv    map[[sha256.Size]byte]db.KV
	order [][sha256.Size]byte
}

func (m *kvMap) Get(k []byte) ([]byte, bool) {
	v, ok := m.kv[sha256.Sum256(k)]
	return v.V, ok
}

func (m *kvMap) Put(k, v []byte) {
	key := sha256.Sum256(k)
	m.kv[key] = db.KV{K: k, V: v}
	m.order = append(m.order, key)
}

// PGTX implements the iden3 transaction interface to use a postgress store
type PGTX struct {
	*PGStore
	cache *kvMap
}

// Get returns the data from a node that is either in the cache or in the db
func (t *PGTX) Get(b []byte) ([]byte, error) {
	fullkey := Concat(t.prefix, b)

	if value, ok := t.cache.Get(fullkey); ok {
		return value, nil
	}

	value, err := t.NodePersister.Get(fullkey)
	if gorm.IsRecordNotFoundError(err) {
		return nil, db.ErrNotFound
	}

	bvalue, err := value.ToDataBytes()

	return bvalue, err
}

// Put adds a new node to the cache
func (t *PGTX) Put(k, v []byte) {
	t.cache.Put(Concat(t.prefix, k[:]), v)
}

// Add copies all nodes from one transaction to this one
func (t *PGTX) Add(atx db.Tx) {
	pgtx := atx.(*PGTX)
	var v db.KV
	for _, key := range pgtx.cache.order {
		v = pgtx.cache.kv[key]
		t.cache.Put(v.K, v.V)
	}
}

// Commit writes all nodes in the cache to the db
func (t *PGTX) Commit() error {
	err := t.NodePersister.Batch(t.cache, t.prefix)
	if err != nil {
		return err
	}
	t.cache.kv = nil
	t.cache.order = nil
	return nil
}

// Close deletes the cache of the transaction
func (t *PGTX) Close() {
	t.cache.kv = nil
	t.cache.order = nil
}
