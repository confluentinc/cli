package httpretry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func withRand(v float64, f func()) {
	old := randFloat
	randFloat = func() float64 { return v }
	defer func() { randFloat = old }()
	f()
}

func respWithRetryAfter(code int, ra string) *http.Response {
	h := http.Header{}
	if ra != "" {
		h.Set("Retry-After", ra)
	}
	return &http.Response{StatusCode: code, Header: h}
}

func TestBackoff_ExponentialAndCap(t *testing.T) {
	withRand(0.5, func() { // 0.5 => zero jitter
		require.Equal(t, 1*time.Second, Backoff(WaitMin, WaitMax, 0, nil))
		require.Equal(t, 1600*time.Millisecond, Backoff(WaitMin, WaitMax, 1, nil))
		require.Equal(t, 2560*time.Millisecond, Backoff(WaitMin, WaitMax, 2, nil))
		require.Equal(t, WaitMax, Backoff(WaitMin, WaitMax, 50, nil))
		require.Equal(t, WaitMax, Backoff(WaitMin, WaitMax, 100000, nil))
	})
}

func TestBackoff_JitterBounds(t *testing.T) {
	withRand(0, func() {
		require.Equal(t, 800*time.Millisecond, Backoff(WaitMin, WaitMax, 0, nil))
	})
	withRand(0.9999999, func() {
		require.InDelta(t, float64(1200*time.Millisecond), float64(Backoff(WaitMin, WaitMax, 0, nil)), float64(time.Millisecond))
		require.LessOrEqual(t, Backoff(WaitMin, WaitMax, 50, nil), WaitMax) // jitter never exceeds the cap
	})
	for i := 0; i < 1000; i++ {
		got := Backoff(WaitMin, WaitMax, 2, nil)
		require.GreaterOrEqual(t, got, time.Duration(float64(2560*time.Millisecond)*0.8)-1)
		require.LessOrEqual(t, got, time.Duration(float64(2560*time.Millisecond)*1.2)+1)
	}
}

func TestBackoff_RetryAfter(t *testing.T) {
	// jitter only extends, never shortens
	withRand(0, func() {
		require.Equal(t, 10*time.Second, Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "10")))
		require.Equal(t, 3*time.Second, Backoff(WaitMin, WaitMax, 4, respWithRetryAfter(503, "3")))
		require.Equal(t, time.Duration(0), Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "0")))
	})
	withRand(0.9999999, func() {
		got := Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "10"))
		require.GreaterOrEqual(t, got, 10*time.Second)
		require.InDelta(t, float64(12*time.Second), float64(got), float64(time.Millisecond))
	})
	// capped at 60s (plus jitter on the capped value)
	withRand(0, func() {
		require.Equal(t, MaxRetryAfter, Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "3600")))
	})
	// unparseable falls back to exponential; other statuses ignore the header
	withRand(0.5, func() {
		require.Equal(t, 1*time.Second, Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "soon")))
		require.Equal(t, 1*time.Second, Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(500, "10")))
	})
}

func TestBackoff_RetryAfterHTTPDate(t *testing.T) {
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return now }
	defer func() { timeNow = time.Now }()
	withRand(0, func() {
		require.Equal(t, 20*time.Second, Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, now.Add(20*time.Second).Format(http.TimeFormat))))
		require.Equal(t, time.Duration(0), Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, now.Add(-time.Hour).Format(http.TimeFormat))))
	})
}

func TestBackoff_NoWait(t *testing.T) {
	NoWait = true
	defer func() { NoWait = false }()
	require.Equal(t, time.Duration(0), Backoff(WaitMin, WaitMax, 3, respWithRetryAfter(429, "30")))
}

func urlErr(method string, err error) error {
	return &url.Error{Op: method, URL: "https://x", Err: err}
}

func TestCheckRetry_Responses(t *testing.T) {
	for _, tt := range []struct {
		method string
		status int
		want   bool
	}{
		{"GET", 200, false}, {"GET", 400, false}, {"GET", 401, false}, {"GET", 404, false},
		{"GET", 429, true}, {"POST", 429, true}, {"PATCH", 429, true},
		{"GET", 503, true}, {"POST", 503, true}, {"PATCH", 503, true},
		{"GET", 500, true}, {"DELETE", 502, true}, {"PUT", 504, true},
		{"POST", 500, false}, {"PATCH", 502, false}, {"POST", 504, false},
		{"GET", 501, false},
	} {
		t.Run(fmt.Sprintf("%s_%d", tt.method, tt.status), func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.status, Request: &http.Request{Method: tt.method}}
			got, err := CheckRetry(context.Background(), resp, nil)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCheckRetry_Errors(t *testing.T) {
	timeoutErr := &net.OpError{Op: "read", Err: timeoutNetErr{}}
	for name, tt := range map[string]struct {
		method string
		err    error
		want   bool
	}{
		"GET timeout":        {"GET", timeoutErr, true},
		"GET reset":          {"GET", &net.OpError{Op: "read", Err: syscall.ECONNRESET}, true},
		"GET EOF":            {"GET", io.EOF, true},
		"GET unexpected EOF": {"GET", io.ErrUnexpectedEOF, true},
		"GET client timeout": {"GET", context.DeadlineExceeded, true},
		"GET temp DNS":       {"GET", &net.DNSError{IsTemporary: true}, true},
		"GET dns not found":  {"GET", &net.DNSError{IsNotFound: true}, false},
		"GET refused":        {"GET", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, false},
		"GET other":          {"GET", errors.New("x509: certificate signed by unknown authority"), false},
		"GET canceled":       {"GET", context.Canceled, false},
		"PUT reset":          {"PUT", io.EOF, true},
		"DELETE reset":       {"DELETE", io.EOF, true},
		"HEAD reset":         {"HEAD", io.EOF, true},
		"OPTIONS reset":      {"OPTIONS", io.EOF, true},
		"POST reset":         {"POST", io.EOF, false},
		"POST timeout":       {"POST", timeoutErr, false},
		"PATCH reset":        {"PATCH", io.EOF, false},
	} {
		t.Run(name, func(t *testing.T) {
			// net/http capitalizes the method like "Get" in url.Error.Op
			op := tt.method[:1] + string([]byte(tt.method[1:]))
			got, err := CheckRetry(context.Background(), nil, urlErr(op, tt.err))
			require.Equal(t, tt.want, got)
			require.Error(t, err) // original error is preserved
		})
	}

	t.Run("unknown method not retried", func(t *testing.T) {
		got, _ := CheckRetry(context.Background(), nil, io.EOF)
		require.False(t, got)
	})
}

type timeoutNetErr struct{}

func (timeoutNetErr) Error() string   { return "timeout" }
func (timeoutNetErr) Timeout() bool   { return true }
func (timeoutNetErr) Temporary() bool { return true }

func TestCheckRetry_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := CheckRetry(ctx, &http.Response{StatusCode: 429}, nil)
	require.False(t, got)
	require.ErrorIs(t, err, context.Canceled)

	ctx, cancel = context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	<-ctx.Done()
	got, err = CheckRetry(ctx, nil, urlErr("Get", io.EOF))
	require.False(t, got)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestErrorHandler(t *testing.T) {
	resp := &http.Response{StatusCode: 429, Body: io.NopCloser(nil)}
	got, err := ErrorHandler(resp, nil, 6)
	require.NoError(t, err)
	require.Same(t, resp, got)

	base := urlErr("Get", io.EOF)
	got, err = ErrorHandler(nil, base, 6)
	require.Nil(t, got)
	require.ErrorIs(t, err, io.EOF)
	require.Contains(t, err.Error(), "gave up after 6 attempts")

	got, err = ErrorHandler(nil, base, 1)
	require.Nil(t, got)
	require.Equal(t, base, err)
}

func newTestClient(opts Options) *http.Client {
	NoWait = true
	c := retryablehttp.NewClient()
	Configure(c, false, opts)
	return c.StandardClient()
}

func TestConfigure_ExhaustionReturnsLastResponse(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"slow down"}`))
	}))
	defer srv.Close()
	defer func() { NoWait = false }()

	resp, err := newTestClient(Options{}).Post(srv.URL, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 429, resp.StatusCode)
	require.Equal(t, "7", resp.Header.Get("Retry-After"))
	body, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(body), "slow down")
	require.EqualValues(t, RetryMax+1, calls.Load())
}

func TestConfigure_RetriesThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer func() { NoWait = false }()

	resp, err := newTestClient(Options{}).Get(srv.URL)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)
	require.EqualValues(t, 3, calls.Load())
}

func TestConfigure_ConnectionErrorRetriedOnlyForIdempotent(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		conn, _, _ := w.(http.Hijacker).Hijack()
		_ = conn.Close() // EOF on the client
	}))
	defer srv.Close()
	defer func() { NoWait = false }()
	client := newTestClient(Options{})

	_, err := client.Get(srv.URL)
	require.Error(t, err)
	require.EqualValues(t, RetryMax+1, calls.Load())

	calls.Store(0)
	_, err = client.Post(srv.URL, "text/plain", nil)
	require.Error(t, err)
	require.EqualValues(t, 1, calls.Load())
}

func TestConfigure_UnauthorizedRefresh(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer func() { NoWait = false }()

	var refreshed atomic.Int32
	resp, err := newTestClient(Options{RefreshOnUnauthorized: func() bool { refreshed.Add(1); return true }}).Get(srv.URL)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)
	require.EqualValues(t, 1, refreshed.Load())

	// without a refresh the 401 is returned untouched
	calls.Store(0)
	resp, err = newTestClient(Options{}).Get(srv.URL)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, 401, resp.StatusCode)
}

func TestConfigure_LongWaitNotified(t *testing.T) {
	old := notifyLongWait
	var notified []time.Duration
	notifyLongWait = func(w time.Duration, _ int) { notified = append(notified, w) }
	defer func() { notifyLongWait = old }()

	c := retryablehttp.NewClient()
	Configure(c, false, Options{})
	withRand(0, func() {
		c.Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "2"))  // short: silent
		c.Backoff(WaitMin, WaitMax, 0, respWithRetryAfter(429, "20")) // long
	})
	require.Equal(t, []time.Duration{20 * time.Second}, notified)
	require.Equal(t, RequestTimeout, c.HTTPClient.Timeout)
	require.Equal(t, RetryMax, c.RetryMax)
}
