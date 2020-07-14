package ctoken

import (
	"time"

	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/internal/token"
	"github.com/overtalk/bgo/pkg/memcache"
)

func init() {
	var module itoken.ITokenModule = new(CTokenModule)
	core.GetCore().RegisterModule(itoken.ModuleName, module)
}

type CTokenModule struct {
	core.Module

	expiredDuration time.Duration
	expiredCallback itoken.ExpiredCB

	cache *memcache.Cache
}

func (tm *CTokenModule) Init() error {
	tm.cache = memcache.NewCache()
	return nil
}

func (tm *CTokenModule) Ticker() (time.Duration, core.Action) {
	if tm.expiredCallback == nil {
		return 0, core.Stop
	}

	for _, userId := range tm.cache.ExpiredKeys(tm.expiredDuration) {
		tm.expiredCallback(userId, tm)
	}
	return tm.expiredDuration / 2, core.Continue
}
