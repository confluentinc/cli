package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// checkRetry decides whether to retry; RetryMax enforces the count. It never returns an error, so a
// response is always handed back untouched. POST and PATCH are retried only on 429 and 503 with Retry-After.
func (cfg Config) checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	st := stateFrom(ctx)
	idempotent := st != nil && st.idempotent

	if ctx.Err() != nil {
		return false, nil
	}

	if err != nil {
		if isTLSAlert(err) {
			// The server rejected the handshake (e.g. our client cert); DefaultRetryPolicy below would retry this case.
			return false, nil
		}
		// Stops on too many redirects, a bad scheme or header, or an untrusted certificate.
		retry, _ := retryablehttp.DefaultRetryPolicy(ctx, nil, err)
		return retry && idempotent, nil
	}

	retryAfter, hasRetryAfter := parseRetryAfter(resp)
	if hasRetryAfter && retryAfter > cfg.RetryWaitMax {
		return false, nil
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return true, nil
	case http.StatusServiceUnavailable:
		return hasRetryAfter || idempotent, nil
	case http.StatusRequestTimeout, http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout:
		return idempotent, nil
	default:
		return false, nil
	}
}

// isTLSAlert reports a handshake the server rejected, for example over the client certificate.
func isTLSAlert(err error) bool {
	opErr, ok := errors.AsType[*net.OpError](err)
	return ok && opErr.Op == "remote error"
}

func parseRetryAfter(resp *http.Response) (time.Duration, bool) {
	if resp == nil || (resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable) {
		return 0, false
	}
	// With negative bounds, DefaultBackoff returns a negative duration unless it parsed Retry-After.
	retryAfter := retryablehttp.DefaultBackoff(-1, -1, 0, resp)
	return retryAfter, retryAfter >= 0
}
