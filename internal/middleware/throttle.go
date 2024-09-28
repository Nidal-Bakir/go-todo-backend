package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

var (
	// errors
	errCapacityExceeded error = errors.New("server capacity exceeded.")
	errTimedOut         error = errors.New("timed out while waiting for a pending request to complete.")
	errContextCanceled  error = errors.New("context was canceled.")

	defaultBacklogTimeout = time.Second * 60
)

// ThrottleOpts represents a set of throttling options.
type ThrottleOpts struct {
	RetryAfterFn   func(ctxDone bool) time.Duration
	Limit          int
	BacklogLimit   int
	BacklogTimeout time.Duration
	StatusCode     int
}

// Throttle is a middleware that limits number of currently processed requests
// at a time across all users. Note: Throttle is not a rate-limiter per user,
// instead it just puts a ceiling on the number of current in-flight requests
// being processed from the point from where the Throttle middleware is mounted.
func Throttle(limit int) func(http.Handler) http.HandlerFunc {
	return ThrottleWithOpts(ThrottleOpts{Limit: limit, BacklogTimeout: defaultBacklogTimeout})
}

// ThrottleBacklog is a middleware that limits number of currently processed
// requests at a time and provides a backlog for holding a finite number of
// pending requests.
func ThrottleBacklog(limit, backlogLimit int, backlogTimeout time.Duration) func(http.Handler) http.HandlerFunc {
	return ThrottleWithOpts(ThrottleOpts{Limit: limit, BacklogLimit: backlogLimit, BacklogTimeout: backlogTimeout})
}

// ThrottleWithOpts is a middleware that limits number of currently processed requests using passed ThrottleOpts.
func ThrottleWithOpts(opts ThrottleOpts) func(http.Handler) http.HandlerFunc {
	utils.Assert(opts.Limit >= 1, "middleware: Throttle expects limit >= 1")
	utils.Assert(opts.BacklogLimit >= 0, "middleware: Throttle expects backlogLimit to be positive")

	statusCode := opts.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusTooManyRequests
	}

	t := throttler{
		tokens:         make(chan token, opts.Limit),
		backlogTokens:  make(chan token, opts.Limit+opts.BacklogLimit),
		backlogTimeout: opts.BacklogTimeout,
		statusCode:     statusCode,
		retryAfterFn:   opts.RetryAfterFn,
	}

	// Filling tokens.
	for i := 0; i < opts.Limit+opts.BacklogLimit; i++ {
		if i < opts.Limit {
			t.tokens <- token{}
		}
		t.backlogTokens <- token{}
	}

	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			select {

			case <-ctx.Done():
				t.setRetryAfterHeaderIfNeeded(w, true)
				utils.WriteError(r.Context(), w, t.statusCode, errContextCanceled)
				return

			case btok := <-t.backlogTokens:
				timer := time.NewTimer(t.backlogTimeout)

				defer func() {
					t.backlogTokens <- btok
				}()

				select {
				case <-timer.C:
					t.setRetryAfterHeaderIfNeeded(w, false)
					utils.WriteError(r.Context(), w, t.statusCode, errTimedOut)
					return
				case <-ctx.Done():
					timer.Stop()
					t.setRetryAfterHeaderIfNeeded(w, true)
					utils.WriteError(r.Context(), w, t.statusCode, errContextCanceled)
					return
				case tok := <-t.tokens:
					defer func() {
						timer.Stop()
						t.tokens <- tok
					}()
					next.ServeHTTP(w, r)
				}
				return

			default:
				t.setRetryAfterHeaderIfNeeded(w, false)
				utils.WriteError(r.Context(), w, t.statusCode, errCapacityExceeded)
				return
			}
		}

	}
}

// token represents a request that is being processed.
type token struct{}

// throttler limits number of currently processed requests at a time.
type throttler struct {
	tokens         chan token
	backlogTokens  chan token
	backlogTimeout time.Duration
	statusCode     int
	retryAfterFn   func(ctxDone bool) time.Duration
}

// setRetryAfterHeaderIfNeeded sets Retry-After HTTP header if corresponding retryAfterFn option of throttler is initialized.
func (t throttler) setRetryAfterHeaderIfNeeded(w http.ResponseWriter, ctxDone bool) {
	if t.retryAfterFn == nil {
		return
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(t.retryAfterFn(ctxDone).Seconds())))
}
