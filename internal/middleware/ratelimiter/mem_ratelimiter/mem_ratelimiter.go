package mem_ratelimiter

import (
	"context"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
)

type memRatelimiter struct {
	conf  ratelimiter.Config
	allow func(ctx context.Context, key string) (bool, time.Duration)
}

func (l *memRatelimiter) Config() ratelimiter.Config {
	return l.conf
}

func (l *memRatelimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	if l.conf.Enabled {
		return l.allow(ctx, l.conf.KeyPrefix+":"+key)
	}
	return true, 0
}
