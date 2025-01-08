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

func NewRedisFixedWindowLimiter(ctx context.Context, redis *redis.Client, conf ratelimiter.Config) ratelimiter.Limiter {
	return &redisRatelimiter{
		conf:  conf,
		redis: redis,
		allow: func(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
			return _fixedWindowAllow(ctx, key, l)
		},
	}
}

func _fixedWindowAllow(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
	perTimeFram := l.conf.RequestsPerTimeFrame
	timeFram := l.conf.TimeFrame
	redisClient := l.redis
	log := zerolog.Ctx(ctx).With().Str("key", key).Int("per_time_fram", perTimeFram).Dur("time_fram", timeFram).Logger()

	keyTimeFram := key + ":" + strconv.FormatInt(time.Now().UnixMilli()/timeFram.Milliseconds(), 10)

	rate, err := redisClient.Get(ctx, keyTimeFram).Int()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Err(err).Msg("Can't rate limit, got an error from redis. Rejecting the request")
			return false, timeFram
		}

		// did not find the key, create new one
		rate = 0
		err = redisClient.Set(ctx, keyTimeFram, rate, timeFram).Err()
		if err != nil {
			log.Err(err).Msg("Can't rate limit, got an error from redis while seting the key value for the first time. Rejecting the request")
			return false, timeFram
		}
	}

	if rate >= perTimeFram {
		remainingTime, err := redisClient.TTL(ctx, keyTimeFram).Result()
		if err != nil {
			log.Err(err).Msg("Can't get the TTL for the key, sending config time fram")
			remainingTime = timeFram
		}
		return false, remainingTime
	}

	err = redisClient.Incr(ctx, keyTimeFram).Err()
	if err != nil {
		log.Err(err).Msg("Can't rate limit, got an error from redis while incr the key value. Rejecting the request")
		return false, timeFram
	}

	return true, 0
}
