package session

import (
	"github.com/overtalk/bgo/pkg/service/packet"
	"github.com/overtalk/bgo/pkg/service/zd"
)

// Request a game request
type Request struct {
	PVer   uint8 // protocol version
	MID    uint8 // module id
	AID    uint8 // action id
	Data   []byte
	Sign   []byte
	buffer zd.IPacketBuffer
}

// GetMID get the mid
func (r *Request) GetMID() uint8 { return r.MID }

// GetAID get the aid
func (r *Request) GetAID() uint8 { return r.AID }

// GetProtoVer get the proto version
func (r *Request) GetProtoVer() uint8 { return r.PVer }

// GetData get the data
func (r *Request) GetData() []byte { return r.Data }

// GetSign get the signature
func (r *Request) GetSign() []byte { return r.Sign }

// Free free its underlying resource
func (r *Request) Free() {
	if r.buffer != nil {
		r.buffer.Free()
	}
}

// NewRequestFromClient create a Request from a client's packet buffer
func NewRequestFromClient(buffer zd.IPacketBuffer) *Request {
	gamePacket := packet.Packet(buffer.Bytes())
	if !gamePacket.IsValid() {
		return nil
	}
	// try to decrypt a game packet using the XORCrypto
	gamePacket.Decrypt(packet.XORCrypto)
	if gamePacket.IsCmdSize() || gamePacket.IsCmdProto() {
		return nil
	}
	var signature []byte
	if gamePacket.HasDataSign() {
		signature = gamePacket.GetDataSign()
	}
	return &Request{
		MID:    gamePacket.GetProtoMID(),
		AID:    gamePacket.GetProtoAID(),
		PVer:   gamePacket.GetProtoVer(),
		Data:   gamePacket.GetDataLoad(),
		Sign:   signature,
		buffer: buffer,
	}
}

// NewRequestFromAgent create a Request from a AgentPacket
func NewRequestFromAgent(pack packet.Packet) *Request {
	var signature []byte
	if pack.HasDataSign() {
		signature = pack.GetDataSign()
	}
	return &Request{
		MID:  pack.GetProtoMID(),
		AID:  pack.GetProtoAID(),
		PVer: pack.GetProtoVer(),
		Data: pack.GetDataLoad(),
		Sign: signature,
	}
}
