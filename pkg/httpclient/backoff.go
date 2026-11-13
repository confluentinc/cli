package httpclient

import (
	"fmt"
	"math"
	"net/http"
	"sync/atomic"
	"time"
)

const noticeThreshold = 3 * time.Second

// Atomic just to be safe: SilenceRetryNotices is meant to be called at startup, before any requests,
// but this keeps a later call safe while requests are running on other goroutines.
var noticesSilenced atomic.Bool

// SilenceRetryNotices turns off the stderr retry notice for the rest of the process.
// The Flink shell and `ai` call it at startup.
func SilenceRetryNotices() {
	noticesSilenced.Store(true)
}

// backoff is called only before a retry that will happen; attempt is 0 before the first retry.
func (cfg Config) backoff(waitMin, waitMax time.Duration, attempt int, resp *http.Response) time.Duration {
	r := cfg.random()

	var wait time.Duration
	if retryAfter, ok := parseRetryAfter(resp); ok {
		// Never shorter than Retry-After; checkRetry has already stopped on values above waitMax.
		wait = retryAfter + time.Duration(r*cfg.Jitter*float64(max(retryAfter, waitMin)))
	} else {
		base := min(float64(waitMin)*math.Pow(growth, float64(attempt)), float64(waitMax))
		wait = min(time.Duration(base*(1+cfg.Jitter*(2*r-1))), waitMax)
	}

	cfg.notice(wait, resp)
	return wait
}

func (cfg Config) notice(wait time.Duration, resp *http.Response) {
	if wait < noticeThreshold || noticesSilenced.Load() {
		return
	}

	reason := "Connection failed"
	if resp != nil {
		reason = fmt.Sprintf("Received %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	cfg.printNotice(fmt.Sprintf("%s; retrying in %s.", reason, wait.Round(time.Second)))
}
