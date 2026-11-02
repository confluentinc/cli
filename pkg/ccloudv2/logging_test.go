package ccloudv2

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCatchLoggingErrorRateLimit(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"12"}}}
	require.EqualError(t, catchLoggingError(nil, resp), "rate limit exceeded for log queries. Retry after 12s")

	// Without the header there is nothing to promise.
	resp = &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	require.EqualError(t, catchLoggingError(nil, resp), "rate limit exceeded for log queries. Try again later")
}

func TestRateLimitError(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	require.EqualError(t, rateLimitError("30", now), "rate limit exceeded for log queries. Retry after 30s")
	require.EqualError(t, rateLimitError(now.Add(90*time.Second).Format(http.TimeFormat), now), "rate limit exceeded for log queries. Retry after 90s")
	require.EqualError(t, rateLimitError(now.Add(-time.Minute).Format(http.TimeFormat), now), "rate limit exceeded for log queries. Try again later")
	require.EqualError(t, rateLimitError("0", now), "rate limit exceeded for log queries. Try again later")
	require.EqualError(t, rateLimitError("soon", now), "rate limit exceeded for log queries. Try again later")
}
