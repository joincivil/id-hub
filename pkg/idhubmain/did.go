package idhubmain

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/utils"
)

var (
	bigCacheConfig = bigcache.Config{
		Shards:             1024,
		LifeWindow:         2 * time.Second,
		CleanWindow:        3 * time.Second,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		Verbose:            true,
		HardMaxCacheSize:   16384,
	}
)

func initEthURIResolver(db *gorm.DB) (*ethuri.Service, error) {
	didPersister := ethuri.NewPostgresPersister(db)
	didService := ethuri.NewService(didPersister)
	return didService, nil
}

func initHTTPUniversalResolver(config *utils.IDHubConfig) (*did.HTTPUniversalResolver, error) {
	bcache, err := bigcache.NewBigCache(bigCacheConfig)
	if err != nil {
		return nil, errors.Wrap(err, "universalresolver.newbigcache")
	}
	cache := did.NewBigCacheResolverCache(bcache)
	resolver := did.NewHTTPUniversalResolver(
		config.DidUniversalResolverHost,
		config.DidUniversalResolverPort,
		cache,
	)
	return resolver, nil
}
