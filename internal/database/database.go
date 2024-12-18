package database

import (
	"context"
	"fmt"

	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Service struct {
	ConnPool *pgxpool.Pool
	Queries  *Queries
}

var (
	database     = os.Getenv("DB_DATABASE")
	password     = os.Getenv("DB_PASSWORD")
	username     = os.Getenv("DB_USERNAME")
	port         = os.Getenv("DB_PORT")
	host         = os.Getenv("DB_HOST")
	poolMaxConns = os.Getenv("DB_POOL_MAX_CONNS")
	dbInstance   *Service
)

func NewConnection(ctx context.Context, log zerolog.Logger) *Service {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	assertEnvVars(log)

	log.Info().Msg(fmt.Sprintf("Connecting to database: %s on port: %s .....", database, port))

	conStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s  sslmode=disable pool_max_conns=%s", username, password, host, port, database, poolMaxConns)
	connectionPool, err := pgxpool.New(ctx, conStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't create new connection to the database")
	}

	err = connectionPool.Ping(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Connection created but can not ping the database")
	}

	log.Info().Msg(fmt.Sprintf("Connected to database: %s on port: %s\n", database, port))

	dbInstance = &Service{
		ConnPool: connectionPool,
		Queries:  New(connectionPool),
	}

	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *Service) Health(ctx context.Context) map[string]string {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	stats := make(map[string]string)

	err := s.ConnPool.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("DB is Down: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.ConnPool.Stat()
	stats["acquired_connections"] = strconv.Itoa(int(dbStats.AcquiredConns()))
	stats["cumulative_acquire_connections"] = strconv.Itoa(int(dbStats.AcquireCount()))
	stats["idle_connections"] = strconv.Itoa(int(dbStats.IdleConns()))
	stats["empty_acquire_count"] = strconv.Itoa(int(dbStats.EmptyAcquireCount()))
	stats["max_conns"] = strconv.Itoa(int(dbStats.MaxConns()))
	stats["max_idle_destroy_count"] = strconv.Itoa(int(dbStats.MaxIdleDestroyCount()))
	stats["max_life_time_destroy_count"] = strconv.Itoa(int(dbStats.MaxLifetimeDestroyCount()))
	stats["acquire_duration"] = strconv.Itoa(int(dbStats.AcquireDuration()))

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
func (s *Service) Close() {
	fmt.Printf("Disconnected from database: %s", database)
	s.ConnPool.Close()
}

func assertEnvVars(log zerolog.Logger) {
	if database == "" {
		log.Fatal().Msg("database env var is empty")
	}
	if password == "" {
		log.Fatal().Msg("password env var is empty")
	}
	if username == "" {
		log.Fatal().Msg("username env var is empty")
	}
	if port == "" {
		log.Fatal().Msg("port env var is empty")
	}
	if host == "" {
		log.Fatal().Msg("host env var is empty")
	}
	if poolMaxConns == "" {
		log.Fatal().Msg("poolMaxConns env var is empty")
	}
}
