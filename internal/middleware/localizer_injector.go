package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Nidal-Bakir/go-todo-backend/internal/l10n"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/resutils"
	"github.com/rs/zerolog"
	"golang.org/x/text/language"
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
			resutils.WriteError(ctx, w, r, http.StatusBadRequest, errors.New("missing Accept-Language in the request header or lang in Query Parameter"))
			return
		}

		tag, err := language.Parse(strings.Split(lang, ",")[0])
		if err != nil {
			resutils.WriteError(ctx, w, r, http.StatusBadRequest, errors.New("invalid Accept-Language header or lang in Query Parameter"))
			return
		}

		baseLang, _ := tag.Base()
		lang = baseLang.String()

		ctx = l10n.ContextWithLocalizer(ctx, l10n.GetLocalizer(lang))
		ctx = zerolog.Ctx(ctx).With().Str("lang", lang).Logger().WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
