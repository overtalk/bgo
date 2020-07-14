package ctcp

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/overtalk/bgo/internal/tcp"
	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/net"
)

func (tcp *CTcpModule) RegisterHandler(handler itcp.HandlerFunc) { tcp.connHandler = handler }

func (tcp *CTcpModule) Start() {
	defer func() {
		if err := recover(); err != nil {
			//logpkg.GetLogger().Error("tcp listen error")
			logpkg.Error("tcp listen error")
		}
	}()

	if tcp.connHandler == nil {
		//logpkg.GetLogger().With(zap.String("reason", "empty tcp handler")).Fatal("failed to start tcp server")
		logpkg.Fatal("failed to start tcp server", zap.String("reason", "empty tcp handler"))
	}

	address := fmt.Sprintf("%s:%d", tcp.cfg.Host, tcp.cfg.Port)
	listener, err := net.Listen(tcp.cfg.Network, address)
	if err != nil {
		//logpkg.GetLogger().With(zap.String("address", address)).Fatal("failed to build tcp listener")
		logpkg.Fatal("failed to build tcp listener", zap.String("address", address))
	}

	// limit listener
	if tcp.cfg.MaxConn > 0 {
		listener = netutil.LimitListener(listener, tcp.cfg.MaxConn)
	}

	//logpkg.GetLogger().With(zap.String("address", address)).Info("start tcp server ")
	logpkg.Info("start tcp server", zap.String("address", address))

	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			// referenced from $GOROOT/src/net/http/server.go:Serve()
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			//logpkg.GetLogger().With(zap.Error(err)).Error("accept error")
			logpkg.Error("accept error", zap.Error(err))
			break
		}
		tempDelay = 0
		// handle every client in its own goroutine
		go tcp.connHandler(conn)
	}
}
