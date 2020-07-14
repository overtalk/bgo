package ipprof

import "github.com/overtalk/bgo/core"

const ModuleName = "internal.pprof"

type IPProf interface {
	core.IModule

	SetPProfPort(ip string, port int)
}
