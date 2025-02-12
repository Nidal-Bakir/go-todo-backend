package redis_ratelimiter

import (
	"context"
	"errors"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewRedisTokenBucketLimiter(ctx context.Context, rdb *redis.Client, conf ratelimiter.Config) ratelimiter.Limiter {
	return &redisRatelimiter{
		conf:             conf,
		rdb:              rdb,
		limiterKeyPrefix: tokenBucketKeyPrefix,
		allow:            _tokenBucketLimiterAllow,
	}
}

func _tokenBucketLimiterAllow(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
	burst := l.conf.PerTimeFrame
	rate := l.conf.TimeFrame
	rdb := l.rdb
	zerolog := zerolog.Ctx(ctx).With().Str("limiter_type", "redis_token_bucket").Str("key", key).Int("burst", burst).Dur("rate", rate).Logger()

	timeTakeToReset := rate * time.Duration(burst)

	pip := rdb.Pipeline()
	
	pip.PTTL(ctx, key)
	pip.GetEx(ctx, key,timeTakeToReset)
		
	pip.Exec(ctx)
	
	curentBurst := 0
	curentBurstStr, err := rdb.Get(ctx, key).Int()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Err(err).Msg("Can't rate limit, got an error from redis while geting the key. Rejecting the request")
			return false, rate
		}
	}

	return true, 0
}
