package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/settings"
)

func settingsRouter(_ context.Context, settingsRepo settings.Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /settings/{label}", readSetting(settingsRepo))
	mux.HandleFunc("POST /settings", setSetting(settingsRepo))
	mux.HandleFunc("DELETE /settings/{label}", deleteSetting(settingsRepo))

	return mux
}

func readSetting(settingsRepo settings.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userAndSession := auth.MustUserAndSessionFromContext(ctx)
		label := r.PathValue("label")
		if label == "" {
			writeError(ctx, w, r, http.StatusBadRequest, fmt.Errorf("invalid label"))
			return
		}

		value, err := settingsRepo.GetSetting(ctx, userAndSession.UserRoleName.String, label)
		if err != nil {
			writeError(ctx, w, r, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}
		writeResponse(ctx, w, r, http.StatusOK, map[string]string{"label": label, "value": value})
	}
}

func setSetting(settingsRepo settings.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		label := r.FormValue("label")
		value := r.FormValue("value")
		if label == "" || value == "" {
			writeError(ctx, w, r, http.StatusBadRequest, fmt.Errorf("invalid label or value"))
			return
		}

		err = settingsRepo.SetSetting(ctx, userAndSession.UserRoleName.String, label, value)
		if err != nil {
			writeError(ctx, w, r, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}
		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
	}
}

func deleteSetting(settingsRepo settings.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		label := r.PathValue("label")
		if label == "" {
			writeError(ctx, w, r, http.StatusBadRequest, fmt.Errorf("invalid label"))
			return
		}

		err = settingsRepo.DeleteSetting(ctx, userAndSession.UserRoleName.String, label)
		if err != nil {
			writeError(ctx, w, r, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}
		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
	}
}
