package did

import (
	"encoding/json"
	"time"

	"github.com/allegro/bigcache"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

var (
	// DefaultBigCacheConfig is a default cache config to use
	DefaultBigCacheConfig = bigcache.Config{
		Shards: 1024,
		// NOTE(PN): Have to test some values here. Most of the time DIDs will not change often
		// but there may be scenario where a DID doc is updated with keys and there is an
		// expectation of immediate usage
		LifeWindow:         5 * time.Minute,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		Verbose:            true,
		HardMaxCacheSize:   16384,
	}
)

// NewBigCacheResolverCache is a convenience function to init and return a new
// BigCacheResolverCache
func NewBigCacheResolverCache(cache *bigcache.BigCache) *BigCacheResolverCache {
	return &BigCacheResolverCache{
		cache: cache,
	}
}

// BigCacheResolverCache implements a ResolverCache using allegro/bigcache
type BigCacheResolverCache struct {
	cache *bigcache.BigCache
}

// Get retrieves the did document out of the cache
func (c *BigCacheResolverCache) Get(d *didlib.DID) (*Document, error) {
	if d == nil {
		return nil, errors.New("did is nil")
	}

	docBys, err := c.cache.Get(d.String())
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil, ErrResolverCacheDIDNotFound
		}

		return nil, errors.Wrap(err, "getfromcache.didcacheget")
	}

	if len(docBys) == 0 {
		return nil, errors.New("Invalid empty doc entry")
	}

	var doc *Document
	err = json.Unmarshal(docBys, &doc)
	if err != nil {
		return nil, errors.Wrap(err, "get.unmarshal")
	}

	return doc, nil
}

// Set sets the did document in the cache given the did
func (c *BigCacheResolverCache) Set(d *didlib.DID, doc *Document) error {
	bys, err := json.Marshal(doc)
	if err != nil {
		return errors.Wrap(err, "set.marshal")
	}

	err = c.cache.Set(d.String(), bys)
	if err != nil {
		return errors.Wrap(err, "set.didcacheset")
	}

	return nil
}
