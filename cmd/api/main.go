package main

import (
	"context"
	"errors"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv" // autoload .env with init function. Do not remove this line
	"github.com/Nidal-Bakir/go-todo-backend/internal/logger"
	"github.com/Nidal-Bakir/go-todo-backend/internal/server"
)

func main() {
	log := logger.NewLogger(AppEnv.IsLocal())

	// Server run context
	serverWithCancelCtx, serverStopCancelFunc := context.WithCancel(context.Background())

	server := server.NewServer(serverWithCancelCtx, log)

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancelFunc := context.WithTimeout(serverWithCancelCtx, 30*time.Second)
		defer shutdownCancelFunc()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Fatal().Msg("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal().Err(err).Msg("Error while shuting down the server.")
		}

		serverStopCancelFunc()
	}()

	log.Info().Msgf("Staring the server on: %s", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("Server Stopped Gracefully.")
		} else {
			log.Fatal().Err(err).Msg("Can't start the server")
		}
	}

	// Wait for server context to be stopped
	<-serverWithCancelCtx.Done()
}
