package itcp

import (
	"net"

	"github.com/overtalk/bgo/core"
)

const ModuleName = "internal.tcp"

type HandlerFunc func(net.Conn)

type ITcpModule interface {
	core.IModule

	Start()
	RegisterHandler(h HandlerFunc)
}
