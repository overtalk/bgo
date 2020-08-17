package packet

import (
	"io"

	"github.com/overtalk/bgo/3rdparty/slab"
	"github.com/pkg/errors"
)

// IPacketBuffer a common packet reading interface, read a raw bytes data,
// and there maybe must different methods to read a packet of data.
type IPacketBuffer interface {
	ReadFrom(io.Reader) (int, error)
	Bytes() []byte
	Free()
}

// A basePacketBuffer with a buffer and valid size
type basePacketBuffer struct {
	data    []byte
	maxSize int
	pool    slab.Pool
}

var _ IPacketBuffer = (*basePacketBuffer)(nil)

// NewPacketBuffer create a IPacketBuffer interface
func NewPacketBuffer(maxSize int, pool slab.Pool) IPacketBuffer {
	return &basePacketBuffer{
		data:    nil,
		maxSize: maxSize,
		pool:    pool,
	}
}

// Alloc get the underlying buffer
func (buf *basePacketBuffer) alloc(size int) {
	if size > buf.maxSize {
		return
	}
	if buf.pool != nil {
		buf.data = buf.pool.Alloc(size)
	} else {
		buf.data = make([]byte, size)
	}
}

// ReadPacket read a packet of data from a Reader
func (buf *basePacketBuffer) ReadFrom(r io.Reader) (int, error) {
	var sizeHeader [2]byte

	// read data length(2 bytes)
	nn, err := io.ReadFull(r, sizeHeader[:2])
	if err != nil {
		return 0, errors.WithMessage(err, "read packet size")
	}
	if nn != 2 {
		return 0, errors.Errorf("read packet size, invalid size(%d!=2)", nn)
	}

	// read payload(N bytes)
	size := int(sizeHeader[0])<<8 + int(sizeHeader[1])
	buf.alloc(2 + size)
	if len(buf.data) == 0 {
		return 0, errors.Errorf("invalid packet size(%d>%d)", size, buf.maxSize-2)
	}

	buf.data[0], buf.data[1] = sizeHeader[0], sizeHeader[1]
	nn, err = io.ReadFull(r, buf.data[2:2+size])
	if err != nil {
		return 0, errors.WithMessagef(err, "read packet(size=%d)", size)
	}

	if nn != size {
		return 0, errors.Errorf("read packet, invalid size(%d!=%d)", nn, size)
	}

	return 2 + nn, nil
}

// Bytes get the real data bytes
func (buf *basePacketBuffer) Bytes() []byte { return buf.data }

// Free release the buffer to its pool
func (buf *basePacketBuffer) Free() {
	if buf.pool != nil && buf.data != nil {
		buf.pool.Free(buf.data)
		buf.data = nil
	}
}
