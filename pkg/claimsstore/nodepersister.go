package claimsstore

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

const (
	// LeafNode is string value signifying node is a leaf
	LeafNode = "leafnode"
	// MiddleNode is string value signifying node is not a leaf
	MiddleNode = "middlenode"
)

var (
	// ErrNoRootCommitForDID is an error for when no root claims are found
	ErrNoRootCommitForDID = errors.New("no rootclaims were in the snapshot")
)

type storageInfo struct {
	KeyCount   int
	ClaimCount int
}

// Node is node in the merkle tree stored in this table
type Node struct {
	gorm.Model
	Prefix       string
	NodeData     string
	NodeKey      string `gorm:"unique_index;not null;"`
	NodeType     string
	DID          string `gorm:"column:did"`
	ClaimType    string
	ClaimVersion uint32
}

// ToDataBytes returns just the data stored in the node as a byte slice
func (c *Node) ToDataBytes() ([]byte, error) {
	return hex.DecodeString(c.NodeData)
}

// ToKV converts the node into the KV type used by iden3
func (c *Node) ToKV() (db.KV, error) {
	prefix := []byte(c.Prefix)
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
		entry, err := merkletree.NewEntryFromBytes(kv.V[1:])
		if err != nil {
			return err
		}
		claimType, version := core.GetClaimTypeVersion(entry)

		if claimType == *claimtypes.ClaimTypeRegisteredDocument ||
			claimType == *claimtypes.ClaimTypeSetRootKeyDID {
			c.DID = hex.EncodeToString(entry.Data[2][:])
		}
		c.ClaimType = hex.EncodeToString(claimType[:])
		c.ClaimVersion = version

		typeName = LeafNode
	}
	c.Prefix = string(prefix)
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

// GetAllForPrefix returns the nodes associated with a particular prefix tree
func (c *NodePGPersister) GetAllForPrefix(prefixBytes []byte, limit int) ([]db.KV, error) {
	prefix := string(prefixBytes)
	var nodes []Node
	var kvs []db.KV
	if err := c.DB.Limit(limit).Order("created_at asc").Where(&Node{Prefix: prefix}).Find(&nodes).Error; err != nil {
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

//GetNextRootClaimVersion gets the next root claim version for a did
func (c *NodePGPersister) GetNextRootClaimVersion(did *didlib.DID) (uint32, error) {
	didbytes, err := claimtypes.HashDID(did)
	if err != nil {
		return 0, err
	}
	node := &Node{}
	if err := c.DB.Where(&Node{
		Prefix:    string(PrefixRootMerkleTree),
		ClaimType: hex.EncodeToString(claimtypes.ClaimTypeSetRootKeyDID[:]),
		DID:       hex.EncodeToString(didbytes),
	}).Order("claim_version desc").Select("claim_version").First(node).Error; err != nil {
		return 0, err
	}
	version := node.ClaimVersion + 1
	return version, nil
}

// GetLatestRootClaimNodes returns the rootclaims nodes for a did sorted by their version number
func (c *NodePGPersister) GetLatestRootClaimNodes(did *didlib.DID) (*[]Node, error) {
	didbytes, err := claimtypes.HashDID(did)
	if err != nil {
		return nil, err
	}
	nodes := &[]Node{}
	if err := c.DB.Where(&Node{
		Prefix:    string(PrefixRootMerkleTree),
		ClaimType: hex.EncodeToString(claimtypes.ClaimTypeSetRootKeyDID[:]),
		DID:       hex.EncodeToString(didbytes),
	}).Order("claim_version desc").Find(nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetLatestRootClaimInSnapshot returns the root claim in the snapshot with the highest version number
func (c *NodePGPersister) GetLatestRootClaimInSnapshot(did *didlib.DID,
	tree *merkletree.MerkleTree) (*claimtypes.ClaimSetRootKeyDID, error) {
	nodes, err := c.GetLatestRootClaimNodes(did)
	if err != nil {
		return nil, err
	}
	for index, node := range *nodes {
		dataBytes, err := hex.DecodeString(node.NodeData)
		// If you get a bad node, then just error
		if err != nil {
			return nil, errors.Wrap(err,
				fmt.Sprintf("GetLatestRootClaimsInSnapshot node at position %v failed to decode from hex", index))
		}
		entry, err := merkletree.NewEntryFromBytes(dataBytes[1:])
		if err != nil {
			return nil, errors.Wrap(err,
				fmt.Sprintf("GetLatestRootClaimsInSnapshot node at position %v failed to create Entry", index))
		}
		_, err = tree.GetDataByIndex(entry.HIndex())
		if err == merkletree.ErrEntryIndexNotFound {
			continue
		} else if err != nil {
			return nil, errors.Wrap(err,
				fmt.Sprintf("GetLatestRootClaimsInSnapshot node at position %v unexpected error retrieving data", index))
		}
		return claimtypes.NewClaimSetRootKeyDIDFromEntry(entry), nil
	}
	return nil, ErrNoRootCommitForDID
}
