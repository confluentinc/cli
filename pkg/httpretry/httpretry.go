// Package httpretry holds the shared retry policy for the CLI's retryable HTTP clients: bounded retries,
// exponential backoff with jitter, Retry-After handling, per-attempt timeouts, and sensible exhaustion behavior.
package httpretry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

const (
	// RetryMax is the maximum number of retries (not counting the initial attempt).
	RetryMax = 5
	// WaitMin is the wait before the first retry.
	WaitMin = 1 * time.Second
	// WaitMax caps any single exponential-backoff wait.
	WaitMax = 30 * time.Second
	// Multiplier is the exponential growth factor between consecutive waits.
	Multiplier = 1.6
	// JitterFraction is the jitter applied to waits (+/-20% for backoff, +0..20% when honoring Retry-After).
	JitterFraction = 0.2
	// MaxRetryAfter caps how long we will honor a server-provided Retry-After. Jitter is added on top of the
	// capped value, so the actual wait can reach MaxRetryAfter*(1+JitterFraction).
	MaxRetryAfter = 60 * time.Second
	// RequestTimeout is the timeout for each individual attempt (set as http.Client.Timeout, so it covers
	// connecting, redirects and reading the response body). It is deliberately not an overall deadline: the
	// total time is bounded by RetryMax and the waits above.
	//
	// Exceptions: these clients are only for API calls. Long transfers such as artifact uploads/downloads
	// (pkg/utils, 20 minute timeout) use their own non-retrying http.Client and are intentionally unaffected.
	RequestTimeout = 30 * time.Second
	// LongWaitThreshold is the wait above which the user is told that the CLI is waiting to retry.
	LongWaitThreshold = 5 * time.Second
)

var (
	// NoWait makes every backoff zero. It is set by the test build of the CLI so that integration tests which
	// exercise retried errors don't sleep.
	NoWait bool

	// randFloat returns a value in [0, 1). Replaceable in tests.
	randFloat = rand.Float64

	// notifyLongWait tells the user about a long wait. Replaceable in tests.
	notifyLongWait = func(wait time.Duration, attempt int) {
		output.ErrPrintf(false, "Waiting %s before retrying the request (retry %d of %d)...\n", wait.Round(time.Second), attempt, RetryMax)
	}

	timeNow = time.Now
)

// Options customizes Configure.
type Options struct {
	// RefreshOnUnauthorized, if set, is called on a 401 response. If it returns true the request is retried.
	RefreshOnUnauthorized func() bool
}

// Configure applies the shared retry policy to client.
func Configure(client *retryablehttp.Client, unsafeTrace bool, opts Options) {
	client.Logger = log.NewLeveledLogger(unsafeTrace)
	client.RetryMax = RetryMax
	client.RetryWaitMin = WaitMin
	client.RetryWaitMax = WaitMax
	client.HTTPClient.Timeout = RequestTimeout
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized && opts.RefreshOnUnauthorized != nil && opts.RefreshOnUnauthorized() {
			return true, nil
		}
		return CheckRetry(ctx, resp, err)
	}
	client.Backoff = func(min, max time.Duration, attempt int, resp *http.Response) time.Duration {
		wait := Backoff(min, max, attempt, resp)
		if wait > LongWaitThreshold {
			notifyLongWait(wait, attempt+1)
		}
		return wait
	}
	client.ErrorHandler = ErrorHandler
}

// CheckRetry decides whether a request should be retried.
//
//   - Context cancellation and deadlines are never retried.
//   - 429 and 503 are retried for every method (the server did not process the request).
//   - Other 5xx responses (500, 502, 504) and connection-level errors (timeouts, resets, EOF, temporary DNS
//     failures) are retried only for idempotent methods, so a POST or PATCH is never silently sent twice.
func CheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if resp == nil || err != nil {
		if err == nil {
			return false, nil
		}
		if errors.Is(err, context.Canceled) {
			return false, err
		}
		// A per-attempt http.Client.Timeout also surfaces as context.DeadlineExceeded; the request context is
		// still live here (checked above), so that case is retryable.
		return isIdempotent(methodFromError(err)) && isTransient(err), err
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return true, nil
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout:
		method := ""
		if resp.Request != nil {
			method = resp.Request.Method
		}
		return isIdempotent(method), nil
	}
	return false, nil
}

func isIdempotent(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false // includes unknown
}

// methodFromError extracts the HTTP method from a *url.Error, which is what net/http returns from Client.Do.
func methodFromError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return strings.ToUpper(urlErr.Op)
	}
	return ""
}

func isTransient(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsTimeout || dnsErr.IsTemporary
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) // per-attempt client timeout
}

// Backoff returns how long to wait before retry number attempt+1 (attempt starts at 0).
//
// Without a usable Retry-After header the wait is min*Multiplier^attempt, capped at max, with +/-JitterFraction
// jitter (re-capped at max). On 429/503 with Retry-After, the wait is the server's value (capped at
// MaxRetryAfter) plus 0..JitterFraction extra (so up to 72s), so we never retry earlier than the server asked.
func Backoff(min, max time.Duration, attempt int, resp *http.Response) time.Duration {
	if NoWait {
		return 0
	}

	if resp != nil && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) {
		if ra, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			if ra > MaxRetryAfter {
				ra = MaxRetryAfter
			}
			return ra + time.Duration(float64(ra)*JitterFraction*randFloat())
		}
	}

	base := float64(min) * math.Pow(Multiplier, float64(attempt))
	if base > float64(max) || math.IsInf(base, 0) {
		base = float64(max)
	}
	jittered := base * (1 + JitterFraction*(2*randFloat()-1))
	if jittered > float64(max) {
		jittered = float64(max)
	}
	return time.Duration(jittered)
}

// parseRetryAfter parses delay-seconds or an HTTP-date.
func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
		if secs < 0 {
			return 0, false
		}
		if secs > int64(MaxRetryAfter/time.Second) {
			return MaxRetryAfter, true
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := t.Sub(timeNow()); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

// ErrorHandler runs when retries are exhausted or a request is not retryable. If there is a response it is
// returned as-is (body unread) so the SDK can build its usual API error from the status and body, rather than an
// opaque "giving up" transport error. Transport errors are returned with the attempt count.
func ErrorHandler(resp *http.Response, err error, attempts int) (*http.Response, error) {
	if resp != nil {
		return resp, err
	}
	if err != nil && attempts > 1 {
		return nil, fmt.Errorf("%w (gave up after %d attempts)", err, attempts)
	}
	return nil, err
}
