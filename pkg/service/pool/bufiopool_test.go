package pool_test

import (
	"bufio"
	"bytes"
	"testing"

	zeropool "github.com/overtalk/bgo/pkg/service/pool"
)

func TestBufReader(t *testing.T) {
	readerPool := zeropool.NewBufReaderPool(100, 1<<10)
	bytesReader := &bytes.Buffer{}
	rdr := readerPool.Get(bytesReader)
	rdr.Free()
}

func BenchmarkBufReaderPool(b *testing.B) {
	readerPool := zeropool.NewBufReaderPool(100, 1<<10)
	for i := 0; i < b.N; i++ {
		bytesReader := &bytes.Buffer{}
		rdr := readerPool.Get(bytesReader)
		rdr.Free()
	}
}

func BenchmarkBufReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytesReader := &bytes.Buffer{}
		bufio.NewReaderSize(bytesReader, 1<<10)
	}
}
