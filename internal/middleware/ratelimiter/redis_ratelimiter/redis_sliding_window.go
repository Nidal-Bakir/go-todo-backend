package redis_ratelimiter

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware/ratelimiter"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func NewRedisSlidingWindowLimiter(ctx context.Context, rdb *redis.Client, conf ratelimiter.Config) ratelimiter.Limiter {
	return &redisRatelimiter{
		conf:             conf,
		rdb:              rdb,
		limiterKeyPrefix: slidingWindowKeyPrefix,
		allow:            _slidingWindowAllow,
	}
}

func _slidingWindowAllow(ctx context.Context, key string, l *redisRatelimiter) (bool, time.Duration) {
	perWindow := l.conf.RequestsPerTimeFrame
	window := l.conf.TimeFrame
	rdb := l.rdb
	zerolog := zerolog.Ctx(ctx).With().Str("key", key).Int("per_window", perWindow).Dur("window", window).Logger()

	rate, err := rdb.HLen(ctx, key).Result()
	if err != nil {
		zerolog.Err(err).Msg("Can't rate limit, unable to get the rate using the HLen on a hash")
		return false, window
	}

	if rate >= int64(perWindow) {
		return false, _calcRemainingTimeForSlidingWindow(ctx, key, window, rdb, zerolog)
	}

	fieldKey := strconv.FormatInt(time.Now().UnixMilli(), 10)
	didAdd, err := rdb.HSetNX(ctx, key, fieldKey, "").Result()
	if err != nil || !didAdd {
		zerolog.Err(err).Msg("Can't rate limit, unable to register new requst in redis hash table")
		return false, window
	}

	err = rdb.HExpire(ctx, key, window, fieldKey).Err()
	if err != nil {
		zerolog.Err(err).Msg("Can't rate limit, unable to set expire for the hash table field")
		return false, window
	}

	return true, 0
}

func _calcRemainingTimeForSlidingWindow(ctx context.Context, key string, window time.Duration, rdb *redis.Client, zerolog zerolog.Logger) time.Duration {
	fields, err := rdb.HKeys(ctx, key).Result()
	if err != nil {
		zerolog.Err(err).Msg("Can't get all fields in order to calc the remaining time, sending window")
		return window
	}

	// if there is not fields that mean the user can request at this moment,
	// and all the Hash fields has been removed due to the fields TTL
	if len(fields) == 0 {
		return 0
	}

	slices.Sort(fields)

	fieldTTL, err := rdb.HPTTL(ctx, key, fields[0]).Result()
	if err != nil {
		zerolog.Err(err).Msg("Can't get the TTL for a hash feild, sending window")
		return window
	}

	// if there is no result then the key must have been expired and removed
	// so the user can call the API at this moment
	if len(fieldTTL) == 0 {
		return 0
	}

	remainingTime := time.Millisecond * time.Duration(fieldTTL[0])
	return remainingTime
}
