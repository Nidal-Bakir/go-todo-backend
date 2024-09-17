package server

import (
	"context"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

func WriteError(ctx context.Context, w http.ResponseWriter, code int, errs ...error) {
	utils.WriteError(ctx, w, code, errs...)
}

func WriteJson(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	utils.WriteJson(ctx, w, code, payload)
}
