package proxy

import (
	"context"
	"net"
)

type LB interface {
	ListenAndServe(ctx context.Context, listener net.Listener) error
}
