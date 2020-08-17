package tunnel

import (
	"github.com/overtalk/bgo/3rdparty/slab"
	"github.com/overtalk/bgo/pkg/service/packet"
	"github.com/overtalk/bgo/pkg/service/pool"
	"github.com/overtalk/bgo/pkg/service/zd"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var backendPool *SessionPool

func InitBackendPool() {
	backendPool = NewSessionPool(
		slab.NewAtomPool(512, 32*1024, 2, 8*1024*1024), // pre-allocated: 56MBytes
		pool.NewBufReaderPool(1000, 64*1024),
	)
}

// BackendRequest a request for backend
type BackendRequest struct {
	buffer zd.IPacketBuffer
}

// NewBackendRequest create a BackendRequest
func NewBackendRequest() *BackendRequest {
	return &BackendRequest{
		buffer: zd.NewPacketBuffer(
			packet.MaxPacketSize,
			backendPool.GetRdrBufPool(),
		),
	}
}

// Read read a steam of data from the backend session
func (req *BackendRequest) Read(r zd.IPacketReader) error {
	return r.ReadPacket(req.buffer)
}

// Free free its underlying resource
func (req *BackendRequest) Free() { req.buffer.Free() }

// GetPacket get the packet
func (req *BackendRequest) GetPacket() packet.Packet { return req.buffer.Bytes() }

// backendConnState a state for backend connection
type backendConnState struct {
	totalTryNum uint32
	lastTryTime int64
}

func newBackendConnState() *backendConnState {
	return &backendConnState{
		totalTryNum: 0,
		lastTryTime: 0,
	}
}

func (bcs *backendConnState) reset() {
	bcs.totalTryNum, bcs.lastTryTime = 0, 0
}

func (bcs *backendConnState) update(now int64) {
	bcs.totalTryNum++
	bcs.lastTryTime = now
}

// binary exponential backoff
var minSecondsForTryConnect = [8]byte{1, 1, 2, 2, 2, 4, 4, 8}

func (bcs *backendConnState) getMinTryTime() int64 {
	return int64(minSecondsForTryConnect[bcs.totalTryNum&0x07])
}

func (bcs *backendConnState) tryAgain() bool {
	nowTS := time.Now().Unix()
	ok := (nowTS - bcs.lastTryTime) >= bcs.getMinTryTime()
	if ok {
		// update its state after each trying
		bcs.update(nowTS)
	}
	return ok
}

// BackendSession backend services
type BackendSession struct {
	id       uint32
	conn     *zd.BaseConn
	closed   int32
	sigClose chan struct{} // notify the session closed
	pingTime int64         // timestamp for ping

	// TODO: use sync.Map to reduce the lock contention
	// manage all FrontendSession attached to it
	lock      sync.RWMutex
	frontends map[uint32]*FrontendSession

	// session id generator
	idCounter uint32
	timeStart time.Time

	// wait all requests being done
	waitRequest *sync.WaitGroup
}

const minPingTime = 20

// NewBackendSession create a BackendSession struct
func NewBackendSession(id uint32, nc net.Conn) *BackendSession {
	baseConn := zd.NewBaseConn(nc, backendPool.GetBufReader(nc))
	baseConn.SetTimeout(10 * time.Second)
	nowTime := time.Now()
	return &BackendSession{
		id:          id,
		conn:        baseConn,
		closed:      0,
		sigClose:    make(chan struct{}),
		pingTime:    nowTime.Unix(),
		frontends:   make(map[uint32]*FrontendSession),
		idCounter:   0,
		timeStart:   nowTime,
		waitRequest: new(sync.WaitGroup),
	}
}

// GetID get the session id
func (this *BackendSession) GetID() uint32 { return this.id }

// ClientAddr get the remote client address
func (this *BackendSession) ClientAddr() string { return this.conn.RemoteAddr() }

// ReadRequest read a request
func (this *BackendSession) ReadRequest() (*BackendRequest, error) {
	req := NewBackendRequest()
	err := req.Read(this.conn)
	return req, err
}

func (this *BackendSession) Write(b []byte) (int, error) { return this.conn.Write(b) }

// Register register it to an agent
func (this *BackendSession) Register(sid uint32) error {
	_, err := this.Write(packet.NewRegister(sid))
	return err
}

// AddRequest add a request to be done
func (this *BackendSession) AddRequest() { this.waitRequest.Add(1) }

// DoneRequest make a request done
func (this *BackendSession) DoneRequest() { this.waitRequest.Done() }

// WaitRequestDone wait all requests done
func (this *BackendSession) WaitRequestDone() { this.waitRequest.Wait() }

// UpdatePing update the ping time
func (this *BackendSession) UpdatePing() { atomic.StoreInt64(&this.pingTime, time.Now().Unix()) }

// Ping send a PingPacket to another endpoint
func (this *BackendSession) Ping() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				// TODO: add log
				//zaplog.S.Error(err)
				//zaplog.S.Error(zap.Stack("").String)
			}
		}()

		ticker := time.NewTicker((minPingTime - 2) * time.Second)
		for {
			select {
			case <-ticker.C:
				//TODO: add log
				//zaplog.S.Infof("ping: agent@%s ---> backend-%d@%s",
				//	s.conn.LocalAddr(), s.id, s.ClientAddr())
				_, err := this.conn.Write(packet.PingPacket)
				if err != nil {
					// TODO: add log
					//zaplog.S.Errorf("ping: agent@%s ---> backend-%d@%s, %v",
					//	s.conn.LocalAddr(), s.id, s.ClientAddr(), err)
					this.conn.Close()
					ticker.Stop()
					return
				}
			case <-this.sigClose:
				ticker.Stop()
				return
			}
		}
	}()
}

// CheckPing check whether the underlying is ok
func (this *BackendSession) CheckPing() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				// TODO: add log
				//zaplog.S.Error(err)
				//zaplog.S.Error(zap.Stack("").String)
			}
		}()

		ticker := time.NewTicker(minPingTime * time.Second)
		for {
			select {
			case now := <-ticker.C:
				lastPingTime := atomic.LoadInt64(&this.pingTime)
				if now.Unix()-lastPingTime > minPingTime {
					// TODO: add log
					// ping timeout
					//zaplog.S.Errorf("ping timeout: agent@%s -> backend-%d@%s",
					//	s.ClientAddr(), s.id, s.conn.LocalAddr())
					this.conn.Close()
					ticker.Stop()
					return
				}
			case <-this.sigClose:
				ticker.Stop()
				return
			}
		}
	}()
}

// NewFrontendSessionID create a tunnel session id
func (this *BackendSession) NewFrontendSessionID() uint32 {
	// id starts from 101, but the returned id may be 0.
	return 100 + atomic.AddUint32(&this.idCounter, 1)
}

// GetFrontendSession add a FrontendSession
func (this *BackendSession) GetFrontendSession(id uint32) *FrontendSession {
	this.lock.RLock()
	sess := this.frontends[id]
	this.lock.RUnlock()
	return sess
}

// AddFrontendSession add a FrontendSession
func (this *BackendSession) AddFrontendSession(sess *FrontendSession) {
	this.lock.Lock()
	this.frontends[sess.GetID()] = sess
	this.lock.Unlock()
}

// DelFrontendSession add a FrontendSession
func (this *BackendSession) DelFrontendSession(id uint32) {
	this.lock.Lock()
	delete(this.frontends, id)
	this.lock.Unlock()
}

// closeAllFrontendSessions close all frontend sessions(not concurrent safely)
func (this *BackendSession) closeAllFrontendSessions() {
	for _, v := range this.frontends {
		if v != nil {
			v.conn.Close()
		}
	}
	// clear all frontend sessions
	this.frontends = map[uint32]*FrontendSession{}
}

// Close close the underlying tcp session and release the resource
func (this *BackendSession) Close() {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		close(this.sigClose)
		this.closeAllFrontendSessions()
		this.conn.Close()
	}
}
