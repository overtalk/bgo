package packet

type IPacket interface {
	GetDataSize() uint16
	SetDataSize(size uint16)
	GetConnID() uint32
	SetConnID(id uint32)
	GetProtoID() uint16
	SetProtoID(id uint16)
	GetProtoVer() uint8
	SetProtoVer(ver uint8)
	GetDataFlag() uint8
	SetDataFlag(flag uint8)
	GetDataSign() []byte
	SetDataSign(sign []byte)
	GetDataLoad() []byte
	SetDataLoad(data []byte)
}
