package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Nidal-Bakir/go-todo-backend/internal/AppEnv" // autoload .env with init function. Do not remove this line
	"github.com/Nidal-Bakir/go-todo-backend/internal/server"
)

func main() {
	server := server.NewServer(context.Background())

	// Server run context
	serverWithCancelCtx, serverStopCancelFunc := context.WithCancel(context.Background())

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
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}

		serverStopCancelFunc()
	}()

	fmt.Println("Staring the server on port: ", os.Getenv("PORT"))
	err := server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("\nServer Stopped Gracefully.")
		} else {
			panic(fmt.Sprintf("can't start the server error: %s", err))
		}
	}

	// Wait for server context to be stopped
	<-serverWithCancelCtx.Done()
}
