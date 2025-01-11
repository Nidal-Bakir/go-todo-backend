package redisdb

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var (
	addr       = os.Getenv("REDIS_ADDR")
	port       = os.Getenv("REDIS_PORT")
	password   = os.Getenv("REDIS_PASSWORD")
	username   = os.Getenv("REDIS_USERNAME")
	clientName = AppEnv.EnvName + "_" + os.Getenv("REDIS_CLIENT_NAME")
	db         = os.Getenv("REDIS_DB")
)

func NewRedisClient(ctx context.Context, log zerolog.Logger) *redis.Client {
	log.Info().Msgf("Connecting to redis server on address=%s, username=%s, clientName=%s .....", addr, username, clientName)

	readTimeout := 50 * time.Second
	client := redis.NewClient(&redis.Options{
		Addr:             addr,
		Network:          "tcp",
		Password:         password,
		Username:         username,
		ClientName:       clientName,
		DB:               utils.Must(strconv.Atoi(db)),
		Protocol:         3,
		ConnMaxIdleTime:  30 * time.Minute,
		DisableIndentity: false,
		PoolFIFO:         false,
		IdentitySuffix:   AppEnv.EnvName,
		MaxActiveConns:   0,
		MaxIdleConns:     1,
		MinIdleConns:     0,
		PoolSize:         10 * runtime.NumCPU(),
		MinRetryBackoff:  8 * time.Millisecond,
		MaxRetryBackoff:  512 * time.Millisecond,
		MaxRetries:       3,
		ReadTimeout:      readTimeout,
		WriteTimeout:     3 * time.Second,
		DialTimeout:      5 * time.Second,
		PoolTimeout:      readTimeout + time.Second,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Info().Msgf("Connected to redis server on address=%s, username=%s, clientName=%s .....", addr, username, clientName)
			return nil
		},
	})

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't PING the redis server")
		return nil
	}

	return client
}
