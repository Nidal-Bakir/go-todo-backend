package server

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/logger"
	"github.com/rs/zerolog"
)

type Server struct {
	port int
	db   *database.Service
	log  zerolog.Logger
}

func NewServer(ctx context.Context) *http.Server {
	log := logger.NewLogger(AppEnv.IsLocal())

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal().Err(err).Msg("Can not read the PORT from env or error while converting to int")
		return nil
	}

	server := &Server{
		port: port,
		db:   database.NewConnection(ctx,log),
		log:  log,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

}
