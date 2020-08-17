package tunnel

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type hostItem struct {
	id   uint32
	host string
}

const (
	serviceON  = 0
	serviceOFF = 1
)

// BackendSessionMgr a manager for BackendSession
type BackendSessionMgr struct {
	//TODO: use sync.Map to reduce the lock contention
	serviceLock  sync.RWMutex
	services     map[uint32]*BackendSession
	serviceState int32 // 0:OK, 1:PAUSE

	// connection controls
	connStateLock sync.RWMutex
	//reg           *registry.ClientRegistry
	buffedHosts chan hostItem
	hosts       atomic.Value // map[uint32]string
	connStates  map[string]*backendConnState
}

func NewBackendSessionMgr() *BackendSessionMgr {
	mgr := &BackendSessionMgr{}
	return mgr
}

var defaultBackendSessionMgr *BackendSessionMgr

// GetBackendSessionMgr get the default backend session mgr
func GetBackendSessionMgr() *BackendSessionMgr {
	return defaultBackendSessionMgr
}

// GetServiceState get the manager's service state
func (mgr *BackendSessionMgr) GetServiceState() int32 {
	return atomic.LoadInt32(&mgr.serviceState)
}

// IsServiceOff check whether the service is off
func (mgr *BackendSessionMgr) IsServiceOff() bool {
	return atomic.LoadInt32(&mgr.serviceState) == serviceOFF
}

// SetServiceState set the manager's service state
func (mgr *BackendSessionMgr) SetServiceState(state int32) {
	atomic.StoreInt32(&mgr.serviceState, state)
}

// SetServiceOn set the manager's service state to ON
func (mgr *BackendSessionMgr) SetServiceOn() {
	atomic.StoreInt32(&mgr.serviceState, serviceON)
}

// SetServiceOff set the manager's service state to OFF
func (mgr *BackendSessionMgr) SetServiceOff() {
	atomic.StoreInt32(&mgr.serviceState, serviceOFF)
}
func (mgr *BackendSessionMgr) getConnState(host string) *backendConnState {
	mgr.connStateLock.Lock()
	state := mgr.connStates[host]
	mgr.connStateLock.Unlock()
	return state
}

func (mgr *BackendSessionMgr) newConnState(host string) *backendConnState {
	mgr.connStateLock.Lock()
	state, flag := mgr.connStates[host]
	if !flag {
		state = newBackendConnState()
		mgr.connStates[host] = state
	}
	mgr.connStateLock.Unlock()
	return state
}

func (mgr *BackendSessionMgr) delConnState(host string) {
	mgr.connStateLock.Lock()
	delete(mgr.connStates, host)
	mgr.connStateLock.Unlock()
}

// get host from inner
func (mgr *BackendSessionMgr) getOneHostFromInner(id uint32) string {
	if hosts, ok := mgr.hosts.Load().(map[uint32]string); ok {
		return hosts[id]
	}
	return ""
}

// get host from register center
func (mgr *BackendSessionMgr) getOneHostFromRegistry(id uint32) string {
	// TODO: get a host from register center
	return ""
}

func (mgr *BackendSessionMgr) getOneHost(id uint32) string {
	if host := mgr.getOneHostFromInner(id); host != "" {
		return host
	}
	return mgr.getOneHostFromRegistry(id)
}

func (mgr *BackendSessionMgr) connectHostAgain(id uint32, host string) {
	select {
	case mgr.buffedHosts <- hostItem{id, host}:
	default:
	}
}

func (mgr *BackendSessionMgr) connectSIDAgain(id uint32) bool {
	// ignore non existed host id
	if host := mgr.getOneHost(id); host != "" {
		// notify dialing to a backend connection
		if mgr.GetSession(id) == nil {
			mgr.connectHostAgain(id, host)
		}
		return true
	}
	//TODO: add log
	//zaplog.S.Errorf("cannot find the host for backend-%d", id)
	return false
}

func (mgr *BackendSessionMgr) dialBackend(item hostItem) {
	if mgr.GetSession(item.id) != nil {
		return
	}
	connState := mgr.getConnState(item.host)
	if connState == nil {
		connState = mgr.newConnState(item.host)
	}
	// dial to a backend connection
	if connState.tryAgain() {
		sess, err := mgr.NewSession(item.id, item.host)
		if err == nil {
			connState.reset()
			// start to handle session request and do ping
			go handleBackendResponse(sess)
		}
	}
}

// only once
func (mgr *BackendSessionMgr) connectHosts() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				//TODO: add log
				//zaplog.S.Error(err)
				//zaplog.S.Error(zap.Stack("").String)
			}
		}()

		for {
			select {
			case item := <-mgr.buffedHosts:
				mgr.dialBackend(item)
			}
			// prevent exhausting cpu resource
			runtime.Gosched()
		}
	}()
}

//////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////
// some func to operate session to backend service
//////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////

// TryGetSession try to get a session by its server id
func (mgr *BackendSessionMgr) TryGetSession(id uint32) *BackendSession {
	for trys := uint32(0); trys < 5; trys++ {
		if sess := mgr.GetSession(id); sess != nil {
			return sess
		}
		if mgr.connectSIDAgain(id) {
			// binary exponential backoff
			time.Sleep((100 * time.Millisecond) * (1 << trys))
		} else {
			// cannot connect this backend id
			return nil
		}
	}
	return nil
}

// NewSession create a session by the host
func (mgr *BackendSessionMgr) NewSession(id uint32, host string) (*BackendSession, error) {
	nc, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err == nil {
		sess := NewBackendSession(id, nc)
		mgr.AddSession(sess)
		return sess, nil
	}
	// TODO: add log
	//zaplog.S.Errorf("dial a backend-%d@%s, %v", id, host, err)
	return nil, err
}

// GetSession get a session by its server id
func (mgr *BackendSessionMgr) GetSession(id uint32) *BackendSession {
	mgr.serviceLock.RLock()
	sess := mgr.services[id]
	mgr.serviceLock.RUnlock()
	return sess
}

// AddSession add a session by its server id
func (mgr *BackendSessionMgr) AddSession(sess *BackendSession) {
	if id := sess.GetID(); id > 0 {
		mgr.serviceLock.Lock()
		mgr.services[id] = sess
		mgr.serviceLock.Unlock()
	} else {
		// TODO: add log
		//zaplog.S.Error("cannot add a backend session, id <= 0")
	}
}

// DelSession del a session by its server id
func (mgr *BackendSessionMgr) DelSession(id uint32) {
	if id > 0 {
		mgr.serviceLock.Lock()
		sess := mgr.services[id]
		delete(mgr.services, id)
		mgr.serviceLock.Unlock()
		// delete conn state
		if sess != nil {
			mgr.delConnState(sess.conn.RemoteAddr())
		}
	} else {
		// TODO: add log
		//zaplog.S.Error("cannot del a backend session, id <= 0")
	}
}
