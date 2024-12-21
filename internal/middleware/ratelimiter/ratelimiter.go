package ratelimiter

import "time"

type Limiter interface {
	Allow(key string) (bool, time.Duration)
	Config() Config
}

type Config struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}
