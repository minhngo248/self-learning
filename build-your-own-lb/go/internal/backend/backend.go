package backend

import "sync/atomic"

type Backend struct {
	addr    string
	nbRetry atomic.Uint32
}

func NewBackend(addr string, nbRetry uint32) *Backend {
	return &Backend{
		addr:    addr,
		nbRetry: atomic.Uint32{},
	}
}

func (be *Backend) Addr() string {
	return be.addr
}

func (be *Backend) NbRetry() uint32 {
	return be.nbRetry.Load()
}

func (be *Backend) SetNbRetry(nbRetry uint32) {
	be.nbRetry.Store(nbRetry)
}

func (be *Backend) IncNbRetry() {
	be.nbRetry.Add(1)
}
