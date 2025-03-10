package server

import (
	"context"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/apiutils"
)

func WriteError(ctx context.Context, w http.ResponseWriter, code int, errs ...error) {
	apiutils.WriteError(ctx, w, code, errs...)
}

func WriteJson(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	apiutils.WriteJson(ctx, w, code, payload)
}
