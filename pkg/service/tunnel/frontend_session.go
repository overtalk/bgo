package tunnel

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/overtalk/bgo/3rdparty/slab"
	"github.com/overtalk/bgo/pkg/service/packet"
	"github.com/overtalk/bgo/pkg/service/pool"
	"github.com/overtalk/bgo/pkg/service/zd"
)

var frontendPool *SessionPool

// InitFrontendPool init some pools for the frontend
func InitFrontendPool() {
	frontendPool = NewSessionPool(
		slab.NewAtomPool(512, 4*1024, 2, 4*1024*1024), // pre-allocated: 16MBytes
		pool.NewBufReaderPool(10000, 1024),
	)
}

type FrontendSession struct {
	id     uint32
	conn   *zd.BaseConn
	buffer zd.IPacketBuffer
	closed int32
	done   chan struct{}

	// connected backend
	backend *BackendSession
}

func NewFrontendSession(nc net.Conn) *FrontendSession {
	baseConn := zd.NewBaseConn(nc, frontendPool.GetBufReader(nc))
	baseConn.SetReadTimeout(10 * time.Second)
	return &FrontendSession{
		id:     0,
		conn:   baseConn,
		buffer: zd.NewPacketBuffer(packet.MaxPacketSize, frontendPool.GetRdrBufPool()),
		done:   make(chan struct{}),
	}
}

func (this *FrontendSession) ClientAddr() string { return this.conn.RemoteAddr() }
func (this *FrontendSession) GetID() uint32      { return this.id }

func (this *FrontendSession) ReadPacket() (packet.Packet, error) {
	if err := this.conn.ReadPacket(this.buffer); err != nil {
		return nil, err
	}

	return this.buffer.Bytes(), nil
}

func (this *FrontendSession) Write(b []byte) (int, error) { return this.conn.Write(b) }

func (this *FrontendSession) Close() {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		this.conn.Close()
		this.buffer.Free()
	}
}

func (this *FrontendSession) IsClosed() bool { return atomic.LoadInt32(&this.closed) == 1 }

func (this *FrontendSession) BindBackendSession(backend *BackendSession) {
	if this.backend == nil {
		this.id = backend.NewFrontendSessionID()
		this.backend = backend
		backend.AddFrontendSession(this)
	}
}

// UnBindBackendSession unbind it from a backend session
func (this *FrontendSession) UnBindBackendSession() {
	if this.backend != nil {
		this.backend.DelFrontendSession(this.id)
		this.id, this.backend = 0, nil
	}
}

// WaitResponse wait its response arriving
func (this *FrontendSession) WaitResponse() {
	select {
	case <-this.done:
	case <-time.After(10 * time.Second):
		//TODO: add log
		//zaplog.S.Errorf("client-%d@%s: response timeout", s.id, s.ClientAddr())
	}
}

// DoneResponse set its completion
func (this *FrontendSession) DoneResponse() { close(this.done) }
