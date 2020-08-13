package zd

import "io"

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
