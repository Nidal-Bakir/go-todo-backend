package mem_ratelimiter

import (
	"context"
	"sync"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"golang.org/x/time/rate"
)

func NewTokenBucketLimiter(ctx context.Context, conf ratelimiter.Config) ratelimiter.Limiter {
	tb := &tokenBucket{
		conf:    conf,
		clients: map[string]*client{},
	}
	tb.startVacuumProc(ctx)

	return &memRatelimiter{
		conf: conf,
		allow: func(ctx context.Context, key string) (bool, time.Duration) {
			return tb.allow(key)
		},
	}
}

type tokenBucket struct {
	conf    ratelimiter.Config
	clients map[string]*client
	mu      sync.Mutex
}

type client struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

func (tb *tokenBucket) allow(key string) (bool, time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	c, ok := tb.clients[key]
	if !ok {
		c = &client{
			lastUsed: time.Now(),
			limiter: rate.NewLimiter(
				rate.Every(tb.conf.TimeFrame),
				tb.conf.PerTimeFrame,
			),
		}
		tb.clients[key] = c
	}
	c.lastUsed = time.Now()
	isAllowed := c.limiter.Allow()
	
	if isAllowed {
		return true, 0
	}
	return false, tb.conf.TimeFrame / time.Duration(tb.conf.PerTimeFrame)
}

func (tb *tokenBucket) startVacuumProc(ctx context.Context) {
	go func() {
	forLoop:
		for {
			time.Sleep(time.Minute)
			select {
			case <-ctx.Done():
				break forLoop
			default:
				tb.mu.Lock()
				for k, v := range tb.clients {
					if time.Since(v.lastUsed) >= time.Minute*3 {
						delete(tb.clients, k)
					}
				}
				tb.mu.Unlock()
			}
		}
	}()
}
