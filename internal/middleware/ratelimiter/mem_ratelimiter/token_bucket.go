package mem_ratelimiter

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
)

func NewTokenBucketLimiter(ctx context.Context, conf ratelimiter.Config) ratelimiter.Limiter {
	tb := &TokenBucket{
		conf:    conf,
		clients: map[string]*client{},
	}
	tb.startVacuumProc(ctx)
	return tb
}

type TokenBucket struct {
	conf    ratelimiter.Config
	clients map[string]*client
	mu      sync.Mutex
}

type client struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

func (tb *TokenBucket) Config() ratelimiter.Config {
	return tb.conf
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, time.Duration) {
	tb.mu.Lock()
	c, ok := tb.clients[key]
	if !ok {
		c = &client{
			lastUsed: time.Now(),
			limiter: rate.NewLimiter(
				rate.Every(tb.conf.TimeFrame),
				tb.conf.RequestsPerTimeFrame,
			),
		}
		tb.clients[key] = c
	}
	c.lastUsed = time.Now()
	isAllowed := c.limiter.Allow()
	tb.mu.Unlock()

	if isAllowed {
		return true, 0
	}
	return false, tb.conf.TimeFrame / time.Duration(tb.conf.RequestsPerTimeFrame)
}

func (tb *TokenBucket) startVacuumProc(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			time.Sleep(time.Minute)
			select {
			case <-ctx.Done():
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
	}(ctx)
}
