package claimsstore

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/jinzhu/gorm"
)

const (
	// LeafNode is string value signifying node is a leaf
	LeafNode = "leafnode"
	// MiddleNode is string value signifying node is not a leaf
	MiddleNode = "middlenode"
)

type storageInfo struct {
	KeyCount   int
	ClaimCount int
}

// Node is node in the merkle tree stored in this table
type Node struct {
	gorm.Model
	DID      string `gorm:"column:did"`
	NodeData string
	NodeKey  string `gorm:"unique_index;not null;"`
	NodeType string
}

// ToDataBytes returns just the data stored in the node as a byte slice
func (c *Node) ToDataBytes() ([]byte, error) {
	return hex.DecodeString(c.NodeData)
}

// ToKV converts the node into the KV type used by iden3
func (c *Node) ToKV() (db.KV, error) {
	prefix := []byte(c.DID)
	key, err := hex.DecodeString(c.NodeKey)
	if err != nil {
		return db.KV{}, err
	}
	value, err := hex.DecodeString(c.NodeData)
	if err != nil {
		return db.KV{}, err
	}
	return db.KV{
		K: key[len(prefix):],
		V: value,
	}, nil
}

// UpdateKVAndPrefix takes a KV and updates the values on the node
func (c *Node) UpdateKVAndPrefix(kv db.KV, prefix []byte) error {
	nodetType := merkletree.NodeType(kv.V[0])
	var typeName string
	if nodetType == merkletree.NodeTypeMiddle {
		typeName = MiddleNode
	} else if nodetType == merkletree.NodeTypeLeaf {
		typeName = LeafNode
	}
	if bytes.Equal(prefix, PrefixRootMerkleTree) || len(prefix) != 32 {
		c.DID = string(prefix)
	} else {
		recoveredDid, err := BinaryToDID(prefix)
		if err != nil {
			return err
		}
		c.DID = recoveredDid.String()
	}
	c.NodeData = hex.EncodeToString(kv.V)
	c.NodeType = typeName
	return nil
}

// TableName sets the name of the corresponding table in the db
func (Node) TableName() string {
	return "claim_nodes"
}

// NodePGPersister is a persister for saving the nodes into postgress
type NodePGPersister struct {
	DB *gorm.DB
}

// NewNodePGPersisterWithDB uses an existing gorm.DB struct to create a new GormPGPersister.
// This is useful if we want to reuse existing connections
func NewNodePGPersisterWithDB(db *gorm.DB) *NodePGPersister {
	gormPGPersister := &NodePGPersister{}
	gormPGPersister.DB = db
	return gormPGPersister
}

// Get a node from the db
func (c *NodePGPersister) Get(key []byte) (*Node, error) {
	strKey := hex.EncodeToString(key)
	node := &Node{}
	if err := c.DB.Find(node, Node{NodeKey: strKey}).Error; err != nil {
		return nil, err
	}

	return node, nil
}

// Batch updates many nodes at once from the cached kv values
// used to update all the middle nodes when a leaf is added or changed
func (c *NodePGPersister) Batch(cache *kvMap, prefix []byte) error {
	tx := c.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var node Node
	var v db.KV
	for _, key := range cache.order {
		v = cache.kv[key]
		node = Node{}
		if err := tx.FirstOrCreate(&node, Node{NodeKey: hex.EncodeToString(v.K)}).Error; err != nil {
			tx.Rollback()
			return err
		}
		err := node.UpdateKVAndPrefix(v, prefix)
		if err != nil {
			return err
		}
		if err := tx.Save(&node).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Info returns basic info about the table
func (c *NodePGPersister) Info() string {
	var totalCount int
	var leafCount int
	if err := c.DB.Table("claim_nodes").Count(&totalCount).Error; err != nil {
		return err.Error()
	}
	if err := c.DB.Model(&Node{}).Where(&Node{NodeType: LeafNode}).Count(&leafCount).Error; err != nil {
		return err.Error()
	}

	json, _ := json.MarshalIndent(
		storageInfo{
			KeyCount:   totalCount,
			ClaimCount: leafCount,
		},
		"", "  ",
	)
	return string(json)
}

// GetAllForDID returns the nodes associated with a particular dids tree
func (c *NodePGPersister) GetAllForDID(didBytes []byte, limit int) ([]db.KV, error) {
	did := string(didBytes)
	var nodes []Node
	var kvs []db.KV
	if err := c.DB.Limit(limit).Order("created_at asc").Where(&Node{DID: did}).Find(&nodes).Error; err != nil {
		return kvs, err
	}
	return convertNodesToKVs(nodes)
}

func convertNodesToKVs(nodes []Node) ([]db.KV, error) {
	var kvs []db.KV
	var kv db.KV
	var err error
	for _, node := range nodes {
		kv, err = node.ToKV()
		if err != nil {
			return kvs, err
		}
		kvs = append(kvs, kv)
	}
	return kvs, nil
}

// GetAll returns all the nodes in all trees
func (c *NodePGPersister) GetAll() ([]db.KV, error) {
	var nodes []Node
	var kvs []db.KV
	if err := c.DB.Find(&nodes).Error; err != nil {
		return kvs, err
	}
	return convertNodesToKVs(nodes)
}
