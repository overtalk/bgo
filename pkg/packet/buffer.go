package packet

import "encoding/binary"

// Buffer defines a packet buffer
type Buffer struct {
	data  []byte
	index int
}

// NewBuffer create a Buffer struct
func NewBuffer(size int) *Buffer { return &Buffer{data: make([]byte, size), index: 0} }

// NewBufferFromBytes create a Buffer struct from a bytes of data
func NewBufferFromBytes(b []byte) *Buffer { return &Buffer{data: b, index: 0} }

// Bytes get all underlying data
func (this *Buffer) Bytes() []byte { return this.data }

// Seek update the data pointer
func (this *Buffer) Seek(offset int) { this.index += offset }

// PutUint8 put a uint8 data
func (this *Buffer) PutUint8(n uint8) {
	this.data[this.index] = n
	this.index++
}

// GetUint8 get a uint8 data
func (this *Buffer) GetUint8() uint8 {
	n := uint8(this.data[this.index])
	this.index++
	return n
}

// PutUint16 put a uint16 data
func (this *Buffer) PutUint16(n uint16) {
	binary.BigEndian.PutUint16(this.data[this.index:this.index+2], n)
	this.index += 2
}

// GetUint16 get a uint16 data
func (this *Buffer) GetUint16() uint16 {
	n := binary.BigEndian.Uint16(this.data[this.index : this.index+2])
	this.index += 2
	return n
}

// PutUint32 put a uint32 data
func (this *Buffer) PutUint32(n uint32) {
	binary.BigEndian.PutUint32(this.data[this.index:this.index+4], n)
	this.index += 4
}

// GetUint32 get a uint32 data
func (this *Buffer) GetUint32() uint32 {
	n := binary.BigEndian.Uint32(this.data[this.index : this.index+4])
	this.index += 4
	return n
}

// PutUint64 put a uint64 data
func (this *Buffer) PutUint64(n uint64) {
	binary.BigEndian.PutUint64(this.data[this.index:this.index+8], n)
	this.index += 8
}

// GetUint64 get a uint64 data
func (this *Buffer) GetUint64() uint64 {
	n := binary.BigEndian.Uint64(this.data[this.index : this.index+8])
	this.index += 8
	return n
}

// PutBytes put several bytes
func (this *Buffer) PutBytes(d []byte) {
	copy(this.data[this.index:], d)
	this.index += len(d)
}

// GetBytes get several bytes
func (this *Buffer) GetBytes(size int) []byte {
	if size == 0 {
		size = len(this.data) - this.index
	}
	d := this.data[this.index : this.index+size]
	this.index += size
	return d
}

// GetAllBytes get all remaining bytes
func (this *Buffer) GetAllBytes() []byte {
	d := this.data[this.index:]
	this.index = len(this.data)
	return d
}
