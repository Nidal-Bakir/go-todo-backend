package ratelimiter

import (
	"time"

	"golang.org/x/time/rate"
)

func NewTokenBucketLimiter(conf Config) Limiter {
	return &TokenBucket{
		conf: conf,
		limiter: rate.NewLimiter(
			rate.Every(conf.TimeFrame),
			conf.RequestsPerTimeFrame,
		),
	}
}

type TokenBucket struct {
	conf    Config
	limiter *rate.Limiter
}

func (tb *TokenBucket) Config() Config {
	return tb.conf
}

func (tb *TokenBucket) Allow(key string) (bool, time.Duration) {
	if tb.limiter.Allow() {
		return true, 0
	}
	return false, tb.conf.TimeFrame / time.Duration(tb.conf.RequestsPerTimeFrame)
}
