package redis_ratelimiter

import (
	"context"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
)

type redisRatelimiter struct {
	conf  ratelimiter.Config
	redis *redis.Client
	allow func(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration)
}

func (l *redisRatelimiter) Config() ratelimiter.Config {
	return l.conf
}

func (l *redisRatelimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	return l.allow(ctx, "tb_rt_"+key, l)
}
