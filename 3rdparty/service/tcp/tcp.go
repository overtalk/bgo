package tcp

import (
	"net"
	"os"
	"time"

	"go.uber.org/zap"

	gozd "github.com/overtalk/bgo/3rdparty/service/zd"
)

// HandlerFunc a Handler wrapper
type HandlerFunc func(net.Conn)

// ListenerOption options for the listener
type ListenerOption struct {
	Name    string      // listener's name
	Network string      // eg: unix/tcp, see net.Dial
	Address string      // eg: socket/ip:port, see net.Dial
	Chmod   os.FileMode // file mode for unix socket, default 0666
	MaxConn int         // listener's maximum connection number
	// if ReadSynced is true, when Listener is closed, all its connections
	// will not read any data. But it may be removed as a default option
	// that it's always stop reading data.
	ReadSynced bool
}

// equal check whether two listener options are equal
func (lo ListenerOption) equal(b *ListenerOption) bool {
	return lo.Network == b.Network && lo.Address == b.Address
}

// ServiceOption a option for a tcp service
type ServiceOption struct {
	Listener ListenerOption
	Handler  HandlerFunc
}

// Service a tcp service
type Service struct {
	opt         ListenerOption
	connHandler HandlerFunc
}

// NewService create a tcp service
func NewService(
	name, addr string, handler HandlerFunc,
) *Service {
	return &Service{
		opt: ListenerOption{
			Name:       name,
			Network:    "tcp",
			Address:    addr,
			MaxConn:    0,
			ReadSynced: false,
		},
		connHandler: handler,
	}
}

// NewServiceWithOption create a tcp service with an option
func NewServiceWithOption(opt ServiceOption) *Service {
	return &Service{
		opt: ListenerOption{
			Name:       opt.Listener.Name,
			Network:    opt.Listener.Network,
			Address:    opt.Listener.Address,
			MaxConn:    opt.Listener.MaxConn,
			ReadSynced: opt.Listener.ReadSynced,
		},
		connHandler: opt.Handler,
	}
}

// GetListenerOption get the listener option
func (ts *Service) GetListenerOption() gozd.ListenerOption {
	return gozd.ListenerOption{
		Address:    ts.opt.Address,
		MaxConn:    ts.opt.MaxConn,
		ReadSynced: ts.opt.ReadSynced,
	}
}

// GetName get the service name
func (ts *Service) GetName() string {
	return ts.opt.Name
}

// NewListener create a service listener
func (ts *Service) NewListener() (net.Listener, error) {
	if ts.opt.Network == "unix" {
		os.Remove(ts.opt.Address)
	}
	l, err := net.Listen(ts.opt.Network, ts.opt.Address)
	if err != nil {
		// handle error
		zap.S().Errorf("bind() failed on: %s %s, error: %v",
			ts.opt.Network, ts.opt.Address, err)
		return nil, err
	}
	if ts.opt.Network == "unix" {
		chmod := ts.opt.Chmod
		if chmod == 0 {
			chmod = 0666
		}
		os.Chmod(ts.opt.Address, chmod)
	}
	return l, nil
}

// Serve service logic
func (ts *Service) Serve(l net.Listener) {
	zap.S().Infof("tcp-service %s: bind to %s@%s",
		ts.opt.Name, ts.opt.Network, ts.opt.Address)

	var tempDelay time.Duration
	for {
		// wait for a network connection
		conn, err := l.Accept()
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
			zap.S().Errorf("accept error: %v", err)
			break
		}
		tempDelay = 0
		// handle every client in its own goroutine
		go ts.connHandler(conn)
	}
}
