package zd

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// LimitListener returns a Listener that accepts at most n simultaneous
// connections from the provided Listener.
// forked from https://github.com/golang/net/blob/master/netutil/listen.go
func LimitListener(l net.Listener, n int) net.Listener {
	return &limitListener{
		Listener: l,
		sem:      make(chan struct{}, n),
		done:     make(chan struct{}),
	}
}

type limitListener struct {
	net.Listener
	// sem 相当于一个队列，没有一个链接进来，就向这个队列中插入一个
	// 后面来的链接就无法链接了
	sem       chan struct{}
	closeOnce sync.Once     // ensures the done chan is only closed once
	done      chan struct{} // no values sent; closed when Close is called
}

// acquire acquires the limiting semaphore. Returns true if successfully
// accquired, false if the listener is closed and the semaphore is not
// acquired.
func (l *limitListener) acquire() bool {
	select {
	case <-l.done:
		return false
	case l.sem <- struct{}{}:
		return true
	}
}

func (l *limitListener) release() { <-l.sem }

func (l *limitListener) Accept() (net.Conn, error) {
	acquired := l.acquire()
	// If the semaphore isn't acquired because the listener was closed, expect
	// that this call to accept won't block, but immediately return an error.
	c, err := l.Listener.Accept()
	if err != nil {
		if acquired {
			l.release()
		}
		return nil, err
	}
	return &limitListenerConn{Conn: c, release: l.release}, nil
}

func (l *limitListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() { close(l.done) })
	return err
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}

// socket types
const (
	unixSocket = "unix"
)

// ListenerOption options for the stoppable listener
type ListenerOption struct {
	Address string // eg: socket/ip:port, see net.Dial
	MaxConn int    // listener's maximum connection number
	// if ReadSynced is true, when Listener is closed, all its connections
	// will not read any data. But it may be removed as a default option
	// that it's always stop reading data.
	ReadSynced bool
}

// A Listener providing a graceful Close process and can be sent
// across processes using the underlying File descriptor.
type Listener interface {
	net.Listener

	// SetDeadline sets the deadline associated with the listener.
	// A zero time value disables the deadline.
	SetDeadline(t time.Time) error
}

// A NetListener support the gracefully shutdown
type NetListener struct {
	Listener

	address string
	maxConn int
	closed  int32

	wg *sync.WaitGroup
	// if readSynced is true, when Listener is closed,
	// all its connections will not read any data.
	readSynced bool
}

// GetAddr get the listener addr
func (nl *NetListener) GetAddr() string {
	return nl.address
}

// GetMaxConn get the max connection limiation
func (nl *NetListener) GetMaxConn() int {
	return nl.maxConn
}

// SetMaxConn get the max connection limiation
func (nl *NetListener) SetMaxConn(n int) {
	nl.maxConn = n
}

// Accept accept a net connection
func (nl *NetListener) Accept() (net.Conn, error) {
	c, err := nl.Listener.Accept()
	if err == nil {
		nl.addConn()
		return &NetListenerConn{c, 0, nl}, nil
	}
	return nil, err
}

// Close close the NetListener
func (nl *NetListener) Close() error {
	err := nl.Listener.Close()
	atomic.StoreInt32(&nl.closed, 1)
	return err
}

// Stop stop the NetListener
func (nl *NetListener) Stop() {
	// http://blog.appsdeck.eu/post/105609534953/graceful-server-restart-with-go
	// After listener timeout, the process still listen on a port, but connections are queued
	// by the network stack of the operating system, waiting for a process to accept them
	nl.SetDeadline(time.Now())
	atomic.StoreInt32(&nl.closed, 1)
}

// isClosed check whether the NetListener is closed
func (nl *NetListener) isClosed() bool {
	return atomic.LoadInt32(&nl.closed) > 0
}

// addConn the listener has accepted a conn
func (nl *NetListener) addConn() {
	nl.wg.Add(1)
}

// delConn a conn form the listener has been done
func (nl *NetListener) delConn() {
	nl.wg.Done()
}

// NetListenerConn allow for us to notice when the connection is closed.
type NetListenerConn struct {
	net.Conn
	closed int32
	nl     *NetListener
}

// Read read data from the connection
func (c *NetListenerConn) Read(b []byte) (int, error) {
	if c.nl.readSynced && c.nl.isClosed() {
		return 0, io.EOF
	}
	return c.Conn.Read(b)
}

// Close close the underlying net connection
func (c *NetListenerConn) Close() (err error) {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		err = c.Conn.Close()
		c.nl.delConn()
	}
	return
}
