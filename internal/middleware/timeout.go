package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/tracker"
	"github.com/rs/zerolog"
)

// Timeout is a middleware that cancels ctx after a given timeout and return
// a 504 Gateway Timeout error to the client.
//
// It's required that you select the ctx.Done() channel to check for the signal
// if the context has reached its deadline and return, otherwise the timeout
// signal will be just ignored.
//
// ie. a route/handler may look like:
//
//	r.Get("/long", func(w http.ResponseWriter, r *http.Request) {
//		ctx := r.Context()
//		processTime := time.Duration(rand.Intn(4)+1) * time.Second
//
//		select {
//		case <-ctx.Done():
//			return
//
//		case <-time.After(processTime):
//			// The above channel simulates some hard work.
//		}
//
//		w.Write([]byte("done"))
//	})
func Timeout(d time.Duration) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctxWithCancel, cancelFunc := context.WithTimeout(r.Context(), d)
			log := *zerolog.Ctx(ctxWithCancel)

			defer func() {
				cancelFunc()
				if errors.Is(ctxWithCancel.Err(), context.DeadlineExceeded) {
					w.WriteHeader(http.StatusGatewayTimeout)

					reqId, ok := tracker.ReqUUIDFromContext(ctxWithCancel)
					logEvent := log.Warn().Err(context.DeadlineExceeded).Int("status_code", http.StatusGatewayTimeout)
					if ok {
						logEvent.Str("request_id", reqId.String())
					}
					logEvent.Msgf("Warning a request timed out. Sending Gateway-Timeout %d status code", http.StatusGatewayTimeout)
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctxWithCancel))
		}
	}
}
