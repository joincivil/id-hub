package claimsstore_test

import (
	"testing"

	"github.com/iden3/go-iden3-core/db"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func setupDBConnection() (*claimsstore.NodePGPersister, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&claimsstore.Node{}).Error
	if err != nil {
		return nil, err
	}

	return claimsstore.NewNodePGPersisterWithDB(db), nil
}

func pgStore(t *testing.T) (db.Storage, *claimsstore.NodePGPersister) {
	persister, err := setupDBConnection()
	if err != nil {
		t.Errorf("failed to create the persister")
	}

	return claimsstore.NewPGStore(persister), persister
}

func testReturnKnownErrIfNotExists(t *testing.T, sto db.Storage) {
	k := []byte("key")

	tx, err := sto.NewTx()
	assert.Nil(t, err)
	_, err = tx.Get(k)
	assert.EqualError(t, err, db.ErrNotFound.Error())
}

func testStorageInsertGet(t *testing.T, sto db.Storage) {
	key := []byte("key")
	value := []byte("data")

	tx, err := sto.NewTx()
	assert.Nil(t, err)
	tx.Put(key, value)
	v, err := tx.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, v)
	assert.Nil(t, tx.Commit())

	tx, err = sto.NewTx()
	assert.Nil(t, err)
	v, err = tx.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, v)
}

func testStorageWithPrefix(t *testing.T, sto db.Storage) {
	k := []byte{9}

	sto1 := sto.WithPrefix([]byte("civil:200"))
	sto2 := sto.WithPrefix([]byte("otherdid:123"))

	// check within tx

	sto1tx, err := sto1.NewTx()
	assert.Nil(t, err)
	sto1tx.Put(k, []byte{4, 5, 6})
	v1, err := sto1tx.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v1, []byte{4, 5, 6})
	assert.Nil(t, sto1tx.Commit())

	sto2tx, err := sto2.NewTx()
	assert.Nil(t, err)
	sto2tx.Put(k, []byte{8, 9})
	v2, err := sto2tx.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v2, []byte{8, 9})
	assert.Nil(t, sto2tx.Commit())

	// check outside tx

	v1, err = sto1.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v1, []byte{4, 5, 6})

	v2, err = sto2.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v2, []byte{8, 9})
}

func testConcatTx(t *testing.T, sto db.Storage) {
	k := []byte{9}

	sto1 := sto.WithPrefix([]byte("civil://1"))
	sto2 := sto.WithPrefix([]byte("civil://2"))

	// check within tx

	sto1tx, _ := sto1.NewTx()
	sto1tx.Put(k, []byte{4, 5, 6})
	sto2tx, _ := sto2.NewTx()
	sto2tx.Put(k, []byte{8, 9})

	sto1tx.Add(sto2tx)
	assert.Nil(t, sto1tx.Commit())

	// check outside tx

	v1, err := sto1.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v1, []byte{4, 5, 6})

	v2, err := sto2.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v2, []byte{8, 9})
}

func testList(t *testing.T, sto db.Storage) {
	sto1 := sto.WithPrefix([]byte("civil:99"))
	r1, err := sto1.List(100)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(r1))

	sto1tx, _ := sto1.NewTx()
	sto1tx.Put([]byte{1}, []byte{4})
	sto1tx.Put([]byte{2}, []byte{5})
	sto1tx.Put([]byte{3}, []byte{6})
	assert.Nil(t, sto1tx.Commit())

	sto2 := sto.WithPrefix([]byte("otherdid:4444"))
	sto2tx, _ := sto2.NewTx()
	sto2tx.Put([]byte{1}, []byte{7})
	sto2tx.Put([]byte{2}, []byte{8})
	sto2tx.Put([]byte{3}, []byte{9})
	assert.Nil(t, sto2tx.Commit())

	r, err := sto1.List(100)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(r))
	assert.Equal(t, db.KV{K: []byte{1}, V: []byte{4}}, r[0])
	assert.Equal(t, db.KV{K: []byte{2}, V: []byte{5}}, r[1])
	assert.Equal(t, db.KV{K: []byte{3}, V: []byte{6}}, r[2])

	r, err = sto1.List(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, db.KV{K: []byte{1}, V: []byte{4}}, r[0])
	assert.Equal(t, db.KV{K: []byte{2}, V: []byte{5}}, r[1])

}

func TestPGStore(t *testing.T) {
	store, persister := pgStore(t)
	defer store.Close()

	cleaner := testutils.DeleteCreatedEntities(persister.DB)
	defer cleaner()

	testReturnKnownErrIfNotExists(t, store)
	testStorageInsertGet(t, store)
	testStorageWithPrefix(t, store)
	testConcatTx(t, store)
	testList(t, store)
}
