package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	ConPool *pgxpool.Pool
	Queries Querier
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

func NewConnection() *Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	assertEnvVars()

	fmt.Printf("Connecting to database: %s on port: %s .....\n", database, port)
	ctx := context.Background()

	conStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s  sslmode=disable pool_max_conns=%s", username, password, host, port, database, poolMaxConns)
	connectionPool, err := pgxpool.New(ctx, conStr)
	if err != nil {
		log.Fatal(err)
	}
	err = connectionPool.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to database: %s on port: %s\n", database, port)
 
	dbInstance = &Service{
		ConPool: connectionPool,
		Queries: New(connectionPool),
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *Service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	stats := make(map[string]string)

	err := s.ConPool.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err))
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.ConPool.Stat()
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
	log.Printf("Disconnected from database: %s", database)
	s.ConPool.Close()
}

func assertEnvVars() {
	if database == "" {
		log.Fatal("database env var is empty")
	}
	if password == "" {
		log.Fatal("password env var is empty")
	}
	if username == "" {
		log.Fatal("username env var is empty")
	}
	if port == "" {
		log.Fatal("port env var is empty")
	}
	if host == "" {
		log.Fatal("host env var is empty")
	}
}
