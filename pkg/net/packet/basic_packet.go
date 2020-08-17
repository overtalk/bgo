package packet

import "encoding/binary"

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
	dataSize := 1 + uint16(len(data))
	packet := NewBasicPacket(dataSize)
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
