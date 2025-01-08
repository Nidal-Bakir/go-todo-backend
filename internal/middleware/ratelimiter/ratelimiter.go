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
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
	KeyPrefix            string
}
