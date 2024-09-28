package middleware

import (
	"net/http"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

func DDD(l int) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		tchan := make(chan Tok, l)

		for range l {
			tchan <- Tok{}
		}

		return func(w http.ResponseWriter, r *http.Request) {

			select {

			case <-r.Context().Done():
				utils.WriteError(r.Context(), w, http.StatusTooManyRequests, errContextCanceled)
				return

			case t := <-tchan:
				defer func() {
					tchan <- t
				}()
				next.ServeHTTP(w, r)

			default:
				utils.WriteError(r.Context(), w, http.StatusTooManyRequests, errCapacityExceeded)
				return
			}

		}
	}
}

type Tok struct{}
