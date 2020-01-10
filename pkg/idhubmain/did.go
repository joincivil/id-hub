package idhubmain

import (
	"github.com/allegro/bigcache"
	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/utils"
)

func initHTTPUniversalResolver(config *utils.IDHubConfig) (*did.HTTPUniversalResolver, error) {
	cacheConfig := bigcache.Config{}
	bcache, err := bigcache.NewBigCache(cacheConfig)
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
