package zd

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// INetServicer a net service
type INetServicer interface {
	GetName() string
	GetListenerOption() ListenerOption
	NewListener() (net.Listener, error)
	Serve(net.Listener)
}

// A NetService supporting stoppable listeners
type NetService struct {
	listener *NetListener // listener instance
	servicer INetServicer // listener handler
	connNum  int32        // current connection number
}

// GetConnNum get the current active conn number
func (ns *NetService) GetConnNum() int32 {
	return atomic.LoadInt32(&ns.connNum)
}

// GetMaxConnNum get the max active conn number
func (ns *NetService) GetMaxConnNum() int {
	return ns.listener.GetMaxConn()
}

// SetMaxConnNum set the max active conn number
func (ns *NetService) SetMaxConnNum(n int) {
	ns.listener.SetMaxConn(n)
}

// RegisterConn register a network connection
func (ns *NetService) RegisterConn() {
	atomic.AddInt32(&ns.connNum, 1)
}

// UnregisterConn unregister a network connection
func (ns *NetService) UnregisterConn() {
	atomic.AddInt32(&ns.connNum, -1)
}

// GetName get the service's name
func (ns *NetService) GetName() string {
	return ns.servicer.GetName()
}

// GetAddr get the service's listen address
func (ns *NetService) GetAddr() string {
	return ns.listener.GetAddr()
}

// CloseListener close the underlying listener
func (ns *NetService) CloseListener() {
	if ns.listener != nil {
		ns.listener.Close()
	}
}

// StopListener stop the underlying listener
func (ns *NetService) StopListener() {
	if ns.listener != nil {
		ns.listener.Stop()
	}
}

// Start start a network service
func (ns *NetService) Start() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
			zap.S().Error(zap.Stack("").String)
		}
	}()
	var l net.Listener = ns.listener
	if maxConn := ns.listener.GetMaxConn(); maxConn > 0 {
		l = LimitListener(l, maxConn)
	}
	ns.servicer.Serve(l)
}

// A NetServer contains several network services
type NetServer struct {
	services    map[string]*NetService
	waitGroup   *sync.WaitGroup
	exitTimeout time.Duration
	exitChan    chan struct{}
	exitOnce    sync.Once
}

const (
	defaultExitTimeout = 3 * time.Minute
)

func newNetServer() *NetServer {
	return &NetServer{
		services:    map[string]*NetService{},
		waitGroup:   new(sync.WaitGroup),
		exitTimeout: defaultExitTimeout,
		exitChan:    make(chan struct{}),
	}
}

// GetService get a network service by its name
func (mgr *NetServer) GetService(name string) *NetService {
	return mgr.services[name]
}

// AddService add a net service
func (mgr *NetServer) AddService(ns INetServicer) {
	mgr.services[ns.GetName()] = &NetService{servicer: ns, connNum: 0}
}

// startService start a network service
func (mgr *NetServer) startService(s *NetService) {
	go func() { s.Start() }()
}

// Run start a network server
func (mgr *NetServer) runServices() {
	for _, v := range mgr.services {
		mgr.startService(v)
	}
}

// initServices init all net services
func (mgr *NetServer) initListeners() error {
	// start listening
	for _, ns := range mgr.services {
		l, err := ns.servicer.NewListener()
		if err != nil {
			return err
		}
		opt := ns.servicer.GetListenerOption()
		ns.listener = &NetListener{
			Listener:   l.(Listener),
			address:    opt.Address,
			maxConn:    opt.MaxConn,
			closed:     0,
			wg:         mgr.waitGroup,
			readSynced: opt.ReadSynced,
		}
	}
	return nil
}

// getExitTimeout get the exit timeout
func (mgr *NetServer) getExitTimeout() time.Duration {
	return mgr.exitTimeout
}

// getExitTimeoutInSecond get the exit timeout in several seconds
func (mgr *NetServer) getExitTimeoutInSecond() int64 {
	return int64(mgr.exitTimeout / time.Second)
}

// setExitTimeout set the exit timeout
func (mgr *NetServer) setExitTimeout(t time.Duration) {
	if t > 0 {
		mgr.exitTimeout = t
	}
}

// stopListeners stop all listeners
func (mgr *NetServer) stopListeners() {
	for _, v := range mgr.services {
		v.StopListener()
	}
}

// exitWaitListeners wait listeners not anymore
func (mgr *NetServer) stopWaitListeners() {
	mgr.exitOnce.Do(func() { close(mgr.exitChan) })
}

// waitListeners wait all listeners exit safely
func (mgr *NetServer) waitListeners() {
	if mgr.exitTimeout > 0 {
		// shutdown process safely
		go func() {
			mgr.waitGroup.Wait()
			mgr.stopWaitListeners()
		}()
		select {
		case <-mgr.exitChan:
		case <-time.After(mgr.exitTimeout):
		}
	}
}
