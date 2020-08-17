package conn

import (
	"net"
	"time"

	"github.com/overtalk/bgo/pkg/net/packet"
	"github.com/overtalk/bgo/pkg/service/pool"
)

// IPacketReader read some data to a IPacketBuffer
type IPacketReader interface {
	ReadPacket(packet.IPacketBuffer) error
}

type BaseConn struct {
	netConn   net.Conn
	bufReader *pool.BufReader

	rdTimeout time.Duration
	wrTimeout time.Duration
}

func NewBaseConn(nc net.Conn, reader *pool.BufReader) *BaseConn {
	return &BaseConn{
		netConn:   nc,
		bufReader: reader,
		rdTimeout: 0,
		wrTimeout: 0,
	}
}

// Read read some bytes from the buffered reader
func (c *BaseConn) Read(b []byte) (int, error) {
	if c.rdTimeout > 0 {
		c.netConn.SetReadDeadline(time.Now().Add(c.rdTimeout))
	}
	return c.bufReader.Read(b)
}

// ReadPacket read the basic packet
func (c *BaseConn) ReadPacket(p packet.IPacketBuffer) (err error) {
	if c.rdTimeout > 0 {
		c.netConn.SetReadDeadline(time.Now().Add(c.rdTimeout))
	}
	_, err = p.ReadFrom(c.bufReader)
	return
}

// Write write some bytes to the wrapped netconn
func (c *BaseConn) Write(b []byte) (int, error) {
	if c.wrTimeout > 0 {
		c.netConn.SetWriteDeadline(time.Now().Add(c.wrTimeout))
	}
	return c.netConn.Write(b)
}

// Close close the wrapped net conn
func (c *BaseConn) Close() (err error) {
	return c.netConn.Close()
}

// SetReadTimeout set the ReadTimeout
func (c *BaseConn) SetReadTimeout(timeout time.Duration) {
	c.rdTimeout = timeout
}

// SetWriteTimeout set the WriteTimeout
func (c *BaseConn) SetWriteTimeout(timeout time.Duration) {
	c.wrTimeout = timeout
}

// SetTimeout set the ReadTimeout and WriteTimeout
func (c *BaseConn) SetTimeout(timeout time.Duration) {
	c.rdTimeout = timeout
	c.wrTimeout = timeout
}

// LocalAddr get the local socket's address string
func (c *BaseConn) LocalAddr() string {
	if addr := c.netConn.LocalAddr(); addr != nil {
		return addr.String()
	}
	return ""
}

// RemoteAddr get the remote socket's address string
func (c *BaseConn) RemoteAddr() string {
	if addr := c.netConn.RemoteAddr(); addr != nil {
		return addr.String()
	}
	return ""
}
