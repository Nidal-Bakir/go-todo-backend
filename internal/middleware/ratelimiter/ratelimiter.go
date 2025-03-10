package ratelimiter

import (
	"context"
	"time"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, time.Duration)
	Config() Config
}

type Config struct {
	PerTimeFrame int
	TimeFrame    time.Duration
	Disabled     bool
	KeyPrefix    string
}
