package server

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	redisdb "github.com/Nidal-Bakir/go-todo-backend/internal/redis_db"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var (
	serverPort = os.Getenv("SERVER_PORT")
)

type Server struct {
	port  int
	db    *database.Service
	redis *redis.Client
	log   zerolog.Logger
}

func NewServer(ctx context.Context, log zerolog.Logger) *http.Server {

	l10n.InitL10n("./l10n", []string{"en", "ar"}, log)

	server := &Server{
		port:  utils.Must(strconv.Atoi(serverPort)),
		db:    database.NewConnection(ctx, log),
		redis: redisdb.NewRedisClient(ctx, log),
		log:   log,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
