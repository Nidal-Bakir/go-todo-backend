package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/appenv"
	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/resutils"
)

func writeError(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, errs ...error) {
	resutils.WriteError(ctx, w, r, code, errs...)
}

func writeResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, payload any) {
	resutils.WriteResponse(ctx, w, r, code, payload)
}

func apiWriteOperationDoneSuccessfullyJson(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	localizer, ok := l10n.LocalizerFromContext(ctx)
	utils.Assert(ok, "we should find the localizer in the context tree, but we did not. something is wrong.")
	writeResponse(ctx, w, r, http.StatusOK, map[string]string{"msg": localizer.GetWithId(l10n.OperationDoneSuccessfullyTrId)})
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

func ReadAuthorizationCookie(r *http.Request) (string, error) {
	authorizationCookie, err := r.Cookie("Authorization")
	if err != nil {
		return "", err
	}
	return authorizationCookie.Value, nil
}

func SetAuthorizationCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		HttpOnly: true,
		Secure:   appenv.IsProdOrStag(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(auth.AuthTokenExpDuration.Seconds()),
	})
}

func RemoveAuthorizationCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		HttpOnly: true,
		Secure:   appenv.IsProdOrStag(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

func StripBearerToken(tokenWithBearer string) string {
	token, _ := strings.CutPrefix(tokenWithBearer, "Bearer ")
	return token
}
