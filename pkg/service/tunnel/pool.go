package tunnel

import (
	"net"

	"github.com/overtalk/bgo/3rdparty/slab"
	"github.com/overtalk/bgo/pkg/service/pool"
)

type SessionPool struct {
	rdrBufPool slab.Pool
	bufRdrPool *pool.BufReaderPool
}

func NewSessionPool(rdrBufPool slab.Pool, bufRdrPool *pool.BufReaderPool) *SessionPool {
	return &SessionPool{
		rdrBufPool: rdrBufPool,
		bufRdrPool: bufRdrPool,
	}
}

// GetBufReader get a bytes buffer from the pool
func (sp *SessionPool) GetBufReader(nc net.Conn) *pool.BufReader {
	return sp.bufRdrPool.Get(nc)
}

// GetRdrBufPool get the rdrBufPool
func (sp *SessionPool) GetRdrBufPool() slab.Pool {
	return sp.rdrBufPool
}
