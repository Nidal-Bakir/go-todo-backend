package resutils

import (
	"context"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/mimes"
)

func WriteError(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, errs ...error) {
	for i, e := range errs {
		if appError := apperr.UnwrapAppErr(e); appError != nil {
			appError.SetTranslation(ctx)
			errs[i] = appError
		}
	}

	if r.Header.Get("Accept") == mimes.App_json {
		apiWriteError(ctx, w, code, errs...)
	} else {
		webWriteError(ctx, w, code, errs...)
	}
}

func WriteResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, payload any) {
	if r.Header.Get("Accept") == mimes.App_json {
		apiWriteJson(ctx, w, code, payload)
	} else {
		// TODO: what?
		panic("WriteResponse with the header Accept != application/json, what shoud we do?")
	}
}
