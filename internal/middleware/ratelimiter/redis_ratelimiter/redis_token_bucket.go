package redis_ratelimiter

import (
	"context"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
)

func NewRedisTokenBucketLimiter(ctx context.Context, redis *redis.Client, conf ratelimiter.Config) ratelimiter.Limiter {
	return &redisRatelimiter{
		conf:  conf,
		redis: redis,
		allow: _tokenBucketLimiterAllow,
	}
}

func _tokenBucketLimiterAllow(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
	return true, 0
}
