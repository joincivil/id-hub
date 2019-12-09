package idhubmain

import (
	"github.com/go-redsync/redsync"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"time"

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
	lockNamespace    = "idhub"
	numAcquireTries  = 256
	retryDelayMillis = 500
)

func initRedisDLock(config *utils.IDHubConfig) lock.DLock {
	// Init pools
	pools := make([]redsync.Pool, len(config.RedisHosts))
	var p redsync.Pool
	var failed bool

	for ind, hp := range config.RedisHosts {
		p = lock.NewRedisDLockPool(hp, poolMaxIdle, poolMaxActive, nil)
		// Check connection
		_, err := redis.DoWithTimeout(p.Get(), 500*time.Millisecond, "PING")
		if err != nil {
			log.Errorf("unable to reach redis host: %v", hp)
			failed = true
			break
		}
		pools[ind] = p
	}

	if !failed {
		log.Infof("Using redis locking")
		dlock := lock.NewRedisDLock(pools, strings.StrToPtr(string(lockNamespace)))
		dlock.MutexTries = numbers.IntToPtr(numAcquireTries)
		dlock.MutexRetryDelayMillis = numbers.IntToPtr(retryDelayMillis)
		return dlock
	}

	log.Errorf("Unable to connect to redis lock pools")
	return nil
}

func initDLock(config *utils.IDHubConfig) lock.DLock {
	// If there are redis hosts in config, use redis dlock
	if config.RedisHosts != nil && len(config.RedisHosts) > 0 {
		dlock := initRedisDLock(config)
		if dlock != nil {
			return dlock
		}
	}

	// Default to local in memory lock
	log.Infof("Using local in-memory locking")
	dlock := lock.NewLocalDLock()
	dlock.Tries = numbers.IntToPtr(numAcquireTries)
	dlock.RetryDelayMillis = numbers.IntToPtr(retryDelayMillis)
	return dlock
}
