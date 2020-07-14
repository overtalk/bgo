package tcppkg

import (
	"encoding/xml"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/net"
	"github.com/overtalk/bgo/utils/xml"
)

type HandlerFunc func(net.Conn)

type Config struct {
	XMLName    xml.Name `xml:"xml"`
	Name       string   `xml:"name"`
	Network    string   `xml:"network"`
	Host       string   `xml:"host"`
	Port       int      `xml:"port"`
	MaxConn    int      `xml:"maxConn"`
	ReadSynced bool     `xml:"readSynced"`
}

type TcpServer struct {
	// config
	cfg *Config
	// some other
	listener    net.Listener
	connHandler HandlerFunc
}

func NewTcpServer(path string) (*TcpServer, error) {
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		return nil, err
	}

	return &TcpServer{cfg: cfg}, nil
}

func (this *TcpServer) RegisterHandler(handler HandlerFunc) { this.connHandler = handler }

func (this *TcpServer) Start() {
	defer func() {
		if err := recover(); err != nil {
			logpkg.Error("this listen error")
		}
	}()

	if this.connHandler == nil {
		logpkg.Fatal("failed to start this server", zap.String("reason", "empty this handler"))
	}

	address := fmt.Sprintf("%s:%d", this.cfg.Host, this.cfg.Port)
	listener, err := net.Listen(this.cfg.Network, address)
	if err != nil {
		logpkg.Fatal("failed to build this listener", zap.String("address", address))
	}

	// limit listener
	if this.cfg.MaxConn > 0 {
		listener = netutil.LimitListener(listener, this.cfg.MaxConn)
	}

	logpkg.Info("start this server", zap.String("address", address))

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

			logpkg.Error("accept error", zap.Error(err))
			break
		}
		tempDelay = 0
		// handle every client in its own goroutine
		go this.connHandler(conn)
	}

}
