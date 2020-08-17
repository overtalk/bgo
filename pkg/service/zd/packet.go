package zd

import (
	"encoding/binary"
	"github.com/overtalk/bgo/3rdparty/slab"
	"github.com/pkg/errors"
	"io"
)

// IPacketReader read some data to a IPacketBuffer
type IPacketReader interface {
	ReadPacket(IPacketBuffer) error
}

// IPacketBuffer a common packet reading interface, read a raw bytes data,
// and there maybe must different methods to read a packet of data.
type IPacketBuffer interface {
	ReadFrom(io.Reader) (int, error)
	Bytes() []byte
	Clone() IPacketBuffer
	Free()
}

//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////

// A basePacketBuffer with a buffer and valid size
type basePacketBuffer struct {
	data    []byte
	maxSize int
	pool    slab.Pool
}

// 确保 basePacketBuffer 实现了 IPacketBuffer 接口
var _ IPacketBuffer = (*basePacketBuffer)(nil)

// NewPacketBuffer create a IPacketBuffer interface
func NewPacketBuffer(maxSize int, pool slab.Pool) IPacketBuffer {
	return &basePacketBuffer{data: nil, maxSize: maxSize, pool: pool}
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
		return 0, errors.Errorf(
			"invalid packet size(%d>%d)", size, buf.maxSize-2,
		)
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
func (buf *basePacketBuffer) Bytes() []byte {
	return buf.data
}

// Clone clone the underlying data excluding the buf field
func (buf *basePacketBuffer) Clone() IPacketBuffer {
	return &basePacketBuffer{data: nil, maxSize: buf.maxSize, pool: buf.pool}
}

// Free release the buffer to its pool
func (buf *basePacketBuffer) Free() {
	if buf.pool != nil && buf.data != nil {
		buf.pool.Free(buf.data)
		buf.data = nil
	}
}

//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////

// BasicPacket 3+N bytes
// datasize(2 bytes) + dataflag(1 byte) + dataload(N bytes)
type BasicPacket []byte

// NewBasicPacket create a BasicPacket
func NewBasicPacket(size uint16) BasicPacket {
	packet := BasicPacket(make([]byte, 2+size))
	packet.SetDataSize(size)
	return packet
}

// MakeBasicPacket make a BasicPacket with data
func MakeBasicPacket(flag uint8, data []byte) BasicPacket {
	datasize := 1 + uint16(len(data))
	packet := NewBasicPacket(datasize)
	packet.SetDataFlag(flag)
	packet.SetDataLoad(data)
	return packet
}

// GetDataSize get the size of packet's payload
func (packet BasicPacket) GetDataSize() uint16 {
	return binary.BigEndian.Uint16(packet[:2])
}

// SetDataSize set the size of packet's payload
func (packet BasicPacket) SetDataSize(datasize uint16) {
	binary.BigEndian.PutUint16(packet[:2], datasize)
}

// GetDataFlag get the data flag
func (packet BasicPacket) GetDataFlag() uint8 {
	return packet[3]
}

// SetDataFlag set the data flag
func (packet BasicPacket) SetDataFlag(flag uint8) {
	packet[3] = flag
}

// GetDataLoad get the packet's dataload
func (packet BasicPacket) GetDataLoad() []byte {
	return packet[3:]
}

// SetDataLoad set the packet's dataload
func (packet BasicPacket) SetDataLoad(data []byte) {
	copy(packet[3:], data)
}

//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////

// TunnelPacket 9+N bytes
// datasize(2 bytes) + client address(6 bytes, ip:port) + dataflag(1 bytes) + dataload(N bytes)
type TunnelPacket []byte

// NewTunnelPacket create a TunnelPacket
func NewTunnelPacket(size uint16) TunnelPacket {
	packet := TunnelPacket(make([]byte, 2+size))
	packet.SetDataSize(size)
	return packet
}

// MakeTunnelPacket make a TunnelPacket with data
func MakeTunnelPacket(addr uint64, flag uint8, data []byte) TunnelPacket {
	packet := NewTunnelPacket(7 + uint16(len(data)))
	packet.SetAddr(addr)
	packet.SetDataFlag(flag)
	packet.SetDataLoad(data)
	return packet
}

// GetDataSize get the size of packet's payload
func (packet TunnelPacket) GetDataSize() uint16 {
	return binary.BigEndian.Uint16(packet[:2])
}

// SetDataSize set the size of packet's payload
func (packet TunnelPacket) SetDataSize(datasize uint16) {
	binary.BigEndian.PutUint16(packet[:2], datasize)
}

// GetAddr get the dataload's address
func (packet TunnelPacket) GetAddr() uint64 {
	u32 := binary.BigEndian.Uint32(packet[2:6])
	u16 := binary.BigEndian.Uint16(packet[6:8])
	return uint64(u32)<<16 | uint64(u16)
}

// SetAddr set the dataload's address
func (packet TunnelPacket) SetAddr(addr uint64) {
	binary.BigEndian.PutUint32(packet[2:6], uint32((addr>>16)&0xFFFFFFFF))
	binary.BigEndian.PutUint16(packet[6:8], uint16(addr&0xFFFF))
}

// GetDataFlag get the data flag
func (packet TunnelPacket) GetDataFlag() uint8 {
	return packet[8]
}

// SetDataFlag set the data flag
func (packet TunnelPacket) SetDataFlag(flag uint8) {
	packet[8] = flag
}

// GetDataLoad get the dataload
func (packet TunnelPacket) GetDataLoad() []byte {
	return packet[9:]
}

// SetDataLoad set the dataload's protocol
func (packet TunnelPacket) SetDataLoad(data []byte) {
	copy(packet[9:], data)
}
