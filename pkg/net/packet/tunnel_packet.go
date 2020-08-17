package packet

import "encoding/binary"

// TunnelPacket 9+N bytes
// datasize(2 bytes) + client address(6 bytes, ip:port) + dataflag(1 bytes) + dataload(N bytes)
type TunnelPacket []byte

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
