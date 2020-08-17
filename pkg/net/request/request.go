package request

import (
	"github.com/overtalk/bgo/pkg/net/packet"
)

// Request a game request
type Request struct {
	PVer   uint8
	MID    uint8
	AID    uint8
	Data   []byte
	Sign   []byte
	buffer packet.IPacketBuffer
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
