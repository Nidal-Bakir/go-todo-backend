package middleware

import (
	"encoding/json"
	"io"

	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var testContent = []byte("Hello world!")

func TestThrottleBacklog(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(time.Second * 1) // Expensive operation.
		w.Write(testContent)
	})

	server := httptest.NewServer(ThrottleBacklog(10, 50, time.Second*10)(mux))
	defer server.Close()

	client := http.Client{
		Timeout: time.Second * 5, // Maximum waiting time.
	}

	var wg sync.WaitGroup

	// The throttler processes 10 consecutive requests, each one of those
	// requests lasts 1s. The maximum number of requests this can possible serve
	// before the clients time out (5s) is 40.
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)

			assertEqual(t, http.StatusOK, res.StatusCode)
			buf, err := io.ReadAll(res.Body)
			assertNoError(t, err)
			assertEqual(t, testContent, buf)
		}(i)
	}

	wg.Wait()
}

func TestThrottleClientTimeout(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(time.Second * 5) // Expensive operation.
		w.Write(testContent)
	})

	server := httptest.NewServer(ThrottleBacklog(10, 50, time.Second*10)(mux))

	defer server.Close()

	client := http.Client{
		Timeout: time.Second * 3, // Maximum waiting time.
	}

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := client.Get(server.URL)
			assertError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestThrottleTriggerGatewayTimeout(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(time.Second * 10) // Expensive operation.
		w.Write(testContent)
	})

	server := httptest.NewServer(ThrottleBacklog(50, 100, time.Second*5)(mux))
	defer server.Close()

	client := http.Client{
		Timeout: time.Second * 60, // Maximum waiting time.
	}

	var wg sync.WaitGroup

	// These requests will be processed normally until they finish.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)
			assertEqual(t, http.StatusOK, res.StatusCode)
		}(i)
	}

	time.Sleep(time.Second * 1)

	// These requests will wait for the first batch to complete but it will take
	// too much time, so they will eventually receive a timeout error.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)

			d := make(map[string]string, 1)
			err = json.NewDecoder(res.Body).Decode(&d)

			assertNoError(t, err)
			assertEqual(t, http.StatusTooManyRequests, res.StatusCode)
			assertEqual(t, errTimedOut.Error(), d["error"])
		}(i)
	}

	wg.Wait()
}

func TestThrottleMaximum(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(time.Second * 3) // Expensive operation.
		w.Write(testContent)
	})

	server := httptest.NewServer(ThrottleBacklog(10, 10, time.Second*5)(mux))

	defer server.Close()

	client := http.Client{
		Timeout: time.Second * 60, // Maximum waiting time.
	}

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)
			assertEqual(t, http.StatusOK, res.StatusCode)

			buf, err := io.ReadAll(res.Body)
			assertNoError(t, err)
			assertEqual(t, testContent, buf)
		}(i)
	}

	// Wait less time than what the server takes to reply.
	time.Sleep(time.Second * 2)

	// At this point the server is still processing, all the following request
	// will be beyond the server capacity.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)

			d := make(map[string]string, 1)
			err = json.NewDecoder(res.Body).Decode(&d)
			assertNoError(t, err)
			assertEqual(t, http.StatusTooManyRequests, res.StatusCode)

			assertEqual(t, errCapacityExceeded.Error(), d["error"])
		}(i)
	}

	wg.Wait()
}

// NOTE: test is disabled as it requires some refactoring. It is prone to intermittent failure.
/*func TestThrottleRetryAfter(t *testing.T) {
	r := chi.NewRouter()

	retryAfterFn := func(ctxDone bool) time.Duration { return time.Hour * 1 }
	r.Use(ThrottleWithOpts(ThrottleOpts{Limit: 10, RetryAfterFn: retryAfterFn}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		time.Sleep(time.Second * 4) // Expensive operation.
		w.Write(testContent)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	client := http.Client{
		Timeout: time.Second * 60, // Maximum waiting time.
	}

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)
			assertEqual(t, http.StatusOK, res.StatusCode)
		}(i)
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Get(server.URL)
			assertNoError(t, err)
			assertEqual(t, http.StatusTooManyRequests, res.StatusCode)
			assertEqual(t, res.Header.Get("Retry-After"), "3600")
		}(i)
	}

	wg.Wait()
}*/

func TestThrottleCustomStatusCode(t *testing.T) {
	const timeout = time.Second * 3

	wait := make(chan struct{})

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-wait:
		case <-time.After(timeout):
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(ThrottleWithOpts(ThrottleOpts{Limit: 1, StatusCode: http.StatusServiceUnavailable})(mux))
	defer server.Close()

	const totalRequestCount = 5

	codes := make(chan int, totalRequestCount)
	errs := make(chan error, totalRequestCount)
	client := &http.Client{Timeout: timeout}
	for i := 0; i < totalRequestCount; i++ {
		go func() {
			resp, err := client.Get(server.URL)
			if err != nil {
				errs <- err
				return
			}
			codes <- resp.StatusCode
		}()
	}

	waitResponse := func(wantCode int) {
		select {
		case err := <-errs:
			t.Fatal(err)
		case code := <-codes:
			assertEqual(t, wantCode, code)
		case <-time.After(timeout):
			t.Fatalf("waiting %d code, timeout exceeded", wantCode)
		}
	}

	for i := 0; i < totalRequestCount-1; i++ {
		waitResponse(http.StatusServiceUnavailable)
	}
	close(wait) // Allow the last request to proceed.
	waitResponse(http.StatusOK)
}
