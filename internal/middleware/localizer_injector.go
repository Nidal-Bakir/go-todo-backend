package middleware

import (
	"errors"
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/rs/zerolog"
)

func LocalizerInjector(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lang := r.Header.Get("Accept-Language")
		langQP := r.FormValue("lang")
		if langQP != "" {
			lang = langQP
		}

		if lang == "" {
			utils.WriteError(r.Context(), w, http.StatusBadRequest, errors.New("missing Accept-Language in Headers or lang in Query Parameter"))
			return
		}

		ctx = l10n.ContextWithLocalizer(ctx, l10n.GetLocalizer(lang))
		ctx = zerolog.Ctx(ctx).With().Str("lang", lang).Logger().WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
