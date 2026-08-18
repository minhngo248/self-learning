package backend

import "sync/atomic"

type Backend struct {
	addr    string
	nbRetry atomic.Uint32
}

func NewBackend(addr string) *Backend {
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

func (be *Backend) ResetNbRetry() {
	be.nbRetry.Store(0)
}

func (be *Backend) IsHealthy() bool {
	return be.NbRetry() == 0
}

func (be *Backend) IsRetrying() bool {
	return be.NbRetry() > 0 && be.NbRetry() <= 3
}

func (be *Backend) IsUnhealthy() bool {
	return be.NbRetry() > 3
}
