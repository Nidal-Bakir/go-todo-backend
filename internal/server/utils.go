package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/apiutils"
)

func writeError(ctx context.Context, w http.ResponseWriter, code int, errs ...error) {
	apiutils.WriteError(ctx, w, code, errs...)
}

func writeJson(ctx context.Context, w http.ResponseWriter, code int, payload any) {
	apiutils.WriteJson(ctx, w, code, payload)
}

func writeOperationDoneSuccessfullyJson(ctx context.Context, w http.ResponseWriter) {
	localizer, ok := l10n.LocalizerFromContext(ctx)
	utils.Assert(ok, "we should find the localizer in the context tree, but we did not. something is wrong.")
	writeJson(ctx, w, http.StatusOK, map[string]string{"msg": localizer.GetWithId(l10n.OperationDoneSuccessfullyTrId)})
}

func return400IfAppErrOr500(err error) int {
	if apperr.IsAppErr(err) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func return400IfApp404IfNoResultErrOr500(err error) int {
	if apperr.IsAppErr(err) {
		if errors.Is(err, apperr.ErrNoResult) {
			return http.StatusNotFound
		}
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
