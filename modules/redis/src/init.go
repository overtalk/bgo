package credis

import (
	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/modules/redis"
	"github.com/overtalk/bgo/pkg/redis"
)

func init() {
	var module iredis.IRedisModule = new(CRedisModule)
	core.GetCore().RegisterModule(iredis.ModuleName, module)
}

type CRedisModule struct {
	core.Module

	redisClient *redispkg.RedisClient
}

func (this *CRedisModule) LoadConfig(path string) error {
	client, err := redispkg.NewRedisClient(path)
	if err != nil {
		return err
	}
	this.redisClient = client
	return nil
}

func (this *CRedisModule) PreTicker() error {
	return this.redisClient.Connect()
}
