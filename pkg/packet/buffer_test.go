package packet_test

import (
	"testing"

	"github.com/overtalk/bgo/pkg/packet"
)

func TestBuffer(t *testing.T) {
	bufferA := packet.NewBuffer(2 + 4 + 8 + 5)
	bufferA.PutUint16(0x1234)
	bufferA.PutUint32(0x12345678)
	bufferA.PutUint64(0x1234567890)
	bufferA.PutBytes([]byte("hello"))
	bufferB := packet.NewBufferFromBytes(bufferA.Bytes())
	bufferB.Seek(2)
	if bufferB.GetUint32() != 0x12345678 {
		t.FailNow()
	}
	bufferB.Seek(-4)
	if bufferB.GetUint16() != 0x1234 {
		t.FailNow()
	}
	bufferB.Seek(-2)
	if bufferB.GetUint16() != 0x1234 {
		t.FailNow()
	}
	if bufferB.GetUint32() != 0x12345678 {
		t.FailNow()
	}
	bufferB.Seek(-4)
	if bufferB.GetUint32() != 0x12345678 {
		t.FailNow()
	}
	if bufferB.GetUint64() != 0x1234567890 {
		t.FailNow()
	}
	if string(bufferB.GetAllBytes()) != "hello" {
		t.FailNow()
	}
}
