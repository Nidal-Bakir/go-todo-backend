package redis_ratelimiter

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func NewRedisFixedWindowLimiter(ctx context.Context, rdb *redis.Client, conf ratelimiter.Config) ratelimiter.Limiter {
	return &redisRatelimiter{
		conf:             conf,
		rdb:              rdb,
		limiterKeyPrefix: fixedWindowKeyPrefix,
		allow:            _fixedWindowAllow,
	}
}

func _fixedWindowAllow(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
	perWindow := l.conf.PerTimeFrame
	window := l.conf.TimeFrame
	redisClient := l.rdb
	log := zerolog.Ctx(ctx).With().Str("limiter_type", "redis_fixed_window").Str("key", key).Int("per_window", perWindow).Dur("window", window).Logger()

	windowKey := key + ":" + strconv.FormatInt(time.Now().UnixMilli()/window.Milliseconds(), 10)

	rate, err := redisClient.Get(ctx, windowKey).Int()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Err(err).Msg("Can't rate limit, got an error from redis while geting the windowKey. Rejecting the request")
			return false, window
		}

		// did not find the key, create new one
		rate = 0
		err = redisClient.Set(ctx, windowKey, rate, window).Err()
		if err != nil {
			log.Err(err).Msg("Can't rate limit, got an error from redis while seting the key value for the first time. Rejecting the request")
			return false, window
		}
	}

	if rate >= perWindow {
		remainingTime, err := redisClient.TTL(ctx, windowKey).Result()
		if err != nil {
			log.Err(err).Msg("Can't get the TTL for the key, sending window")
			remainingTime = window
		}
		return false, remainingTime
	}

	err = redisClient.Incr(ctx, windowKey).Err()
	if err != nil {
		log.Err(err).Msg("Can't rate limit, got an error from redis while incr the key value. Rejecting the request")
		return false, window
	}

	return true, 0
}
