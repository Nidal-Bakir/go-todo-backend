package redis_ratelimiter

import (
	"context"
	"fmt"
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
	log := zerolog.Ctx(ctx).With().Str("key", key).Int("per_window", perWindow).Dur("window", window).Logger()

	currentTimeMicro := time.Now().UnixMicro()
	windowStartTimeMicro := currentTimeMicro - window.Microseconds()

	pip := rdb.TxPipeline()

	pip.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(windowStartTimeMicro))
	pip.ZCard(ctx, key)

	pipRes, err := pip.Exec(ctx)
	if err != nil {
		log.Err(err).Msg("Can't rate limit, got an error from redis while exec the TxPipeline. Rejecting the request")
		return false, window
	}

	redisIntCmd, ok := pipRes[len(pipRes)-1].(*redis.IntCmd)
	if !ok {
		log.Err(err).Msg("Can't cast the Pip result Cmder to IntCmd, cast error. Rejecting the request")
		return false, window
	}

	rate, err := redisIntCmd.Result()
	if err != nil {
		log.Err(err).Msg("Can't rate limit, unable to get the rate from the ZCard from TxPipeline result. Rejecting the request")
		return false, window
	}

	if rate >= int64(perWindow) {
		resultSetData, err := rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
			Key:     key,
			Start:   "-inf",
			Stop:    "+inf",
			ByScore: true,
			Offset:  0,
			Count:   1,
		}).Result()

		if err != nil || len(resultSetData) == 0 {
			if err != nil {
				log.Err(err).Msg("Can't calc the remaining time, sending window")
			}
			return false, window
		}

		remainingTime := int64(resultSetData[0].Score) - windowStartTimeMicro
		remainingTimeMicro := time.Microsecond * time.Duration(remainingTime)
		return false, remainingTimeMicro
	}

	pip = rdb.TxPipeline()
	pip.ZAdd(ctx, key, redis.Z{Score: float64(currentTimeMicro), Member: fmt.Sprint(currentTimeMicro)})
	pip.Expire(ctx, key, window)
	pipRes, err = pip.Exec(ctx)
	if err != nil {
		log.Err(err).Msg("Can't rate limit, got an error from redis while exec the TxPipeline. Rejecting the request")
		return false, window
	}

	return true, 0
}
