package redis_ratelimiter

import (
	"context"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
)

func NewRedisLeakyBucketLimiter(ctx context.Context, rdb *redis.Client, conf ratelimiter.Config) ratelimiter.Limiter {
	return &redisRatelimiter{
		conf:             conf,
		rdb:              rdb,
		limiterKeyPrefix: leakyBuketKeyPrefix,
		allow:            _leakyBucketLimiterAllow,
	}
}

func _leakyBucketLimiterAllow(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
	// perTimeFram := l.conf.RequestsPerTimeFrame
	// timeFram := l.conf.TimeFrame
	return true, 0
}
