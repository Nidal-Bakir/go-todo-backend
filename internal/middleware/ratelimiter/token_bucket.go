package ratelimiter

import (
	"time"
)

func NewTokenBucketLimiter(conf Config) Limiter {
	return &TokenBucket{conf: conf}
}

type TokenBucket struct {
	conf Config
}

func (tb *TokenBucket) Allow(key string) (bool, time.Duration) {
	return false, time.Second * 5
}

func (tb *TokenBucket) Config() Config {
	return tb.conf
}
