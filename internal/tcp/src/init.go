package ctcp

import (
	"net"

	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/internal/tcp"
)

func init() {
	var module itcp.ITcpModule = new(CTcpModule)
	core.GetCore().RegisterModule(itcp.ModuleName, module)
}

type CTcpModule struct {
	core.Module

	// config & other modules
	cfg *Config
	// some other
	listener    net.Listener
	connHandler itcp.HandlerFunc
}

func (tcp *CTcpModule) PreTicker() error {
	go tcp.Start()
	return nil
}
