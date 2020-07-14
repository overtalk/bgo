package iredis

import "github.com/overtalk/bgo/core"

const ModuleName = "modules.redis"

type IRedisModule interface {
	core.IModule
}
