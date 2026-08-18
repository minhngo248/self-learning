package health

import (
	"context"
	"sync"
)

type HealthChecker interface {
	CheckHealth(ctx context.Context, wg *sync.WaitGroup)
}
