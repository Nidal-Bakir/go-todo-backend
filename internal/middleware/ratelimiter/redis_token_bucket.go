package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisTokenBucketLimiter(ctx context.Context, redis *redis.Client, conf Config) Limiter {
	return &RedisTokenBucket{
		conf:  conf,
		redis: redis,
	}
}

type RedisTokenBucket struct {
	conf  Config
	redis *redis.Client
}

func (tb *RedisTokenBucket) Config() Config {
	return tb.conf
}

func (tb *RedisTokenBucket) Allow(ctx context.Context, key string) (bool, time.Duration) {
	return tb.allow(ctx, "rate_limiter_"+key)
}

func (tb *RedisTokenBucket) allow(ctx context.Context, key string) (bool, time.Duration) {
	return true, 0
}
