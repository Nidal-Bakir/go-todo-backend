package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/Nidal-Bakir/go-todo-backend/internal/app_env" // autoload .env with init function. Do not remove this line
	"github.com/Nidal-Bakir/go-todo-backend/internal/server"
)

func main() {
	server := server.NewServer(context.Background())

	fmt.Println("Staring the server on port: ", os.Getenv("PORT"))
	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
