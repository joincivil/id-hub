package idhubmain

import (
	"github.com/go-redsync/redsync"

	"github.com/joincivil/go-common/pkg/lock"
	"github.com/joincivil/go-common/pkg/numbers"
	"github.com/joincivil/go-common/pkg/strings"

	"github.com/joincivil/id-hub/pkg/utils"
)

var (
	poolMaxActive = numbers.IntToPtr(4)
	poolMaxIdle   = numbers.IntToPtr(2)
)

const (
	lockNamespace = "idhub"
)

func initDLock(config *utils.IDHubConfig) lock.DLock {
	// Init pools
	pools := make([]redsync.Pool, len(config.RedisHosts))
	for ind, hp := range config.RedisHosts {
		pools[ind] = lock.NewRedisDLockPool(hp, poolMaxIdle, poolMaxActive, nil)
	}

	return lock.NewRedisDLock(pools, strings.StrToPtr(string(lockNamespace)))
}
