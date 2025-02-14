package redis_ratelimiter

import (
	"context"
	"errors"

	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
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
	zlog := zerolog.Ctx(ctx).With().Str("limiter_type", "redis_token_bucket").Str("key", key).Int("burst", burst).Dur("rate", rate).Logger()

	timeTakeToReset := rate * time.Duration(burst)

	pip := rdb.Pipeline()
	pip.PTTL(ctx, key) // 0 for TTL
	pip.Get(ctx, key)  // 1 for GET
	pipRes, err := pip.Exec(ctx)

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			zlog.Err(err).Msg("Can't rate limit, got an error from redis while geting the key. Rejecting the request")
			return false, rate
		}
	}

	ttl, err := pipRes[0].(*redis.DurationCmd).Result()
	if err != nil {
		zlog.Err(err).Msg("Can't rate limit, got an error from redis while geting the key PTTL. Rejecting the request")
		return false, rate
	}
	if ttl < 0 { // if the ttl is in negative range that means there is no ttl on that key
		ttl = 0
	}

	curentBurst, err := pipRes[1].(*redis.StringCmd).Int()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			zlog.Err(err).Msg("Can't rate limit, got an error from redis while geting the current burst. Rejecting the request")
			return false, rate
		}
	}

	curentBurst = fillBucket(curentBurst, ttl, timeTakeToReset, rate)

	if curentBurst >= burst {
		retryAfter := rate - (timeTakeToReset - ttl)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	curentBurst++ // the current request

	err = rdb.SetEx(ctx, key, curentBurst, timeTakeToReset).Err()
	if err != nil {
		zlog.Err(err).Msg("Can't rate limit, got an error from redis while seting the curent burst. Rejecting the request")
		return false, rate
	}

	return true, 0
}

func fillBucket(curentBurst int, ttl, timeTakeToReset, rate time.Duration) int {
	if curentBurst == 0 || ttl == 0 {
		return 0
	}

	lastUpdated := time.Now().Add(-(timeTakeToReset - ttl))
	gainedTokens := time.Since(lastUpdated).Milliseconds() / rate.Milliseconds()

	newCurentBurst := curentBurst - int(gainedTokens)
	if newCurentBurst < 0 {
		newCurentBurst = 0
	}

	return newCurentBurst
}
