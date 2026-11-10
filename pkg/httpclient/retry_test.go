package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func contextWithState(idempotent bool) context.Context {
	return context.WithValue(context.Background(), stateKey{}, &state{idempotent: idempotent})
}

func response(status int, retryAfter string) *http.Response {
	resp := &http.Response{StatusCode: status, Header: http.Header{}}
	if retryAfter != "" {
		resp.Header.Set("Retry-After", retryAfter)
	}
	return resp
}

func TestCheckRetry_Responses(t *testing.T) {
	tests := []struct {
		name               string
		resp               *http.Response
		retryIdempotent    bool
		retryNonIdempotent bool
	}{
		{"429", response(http.StatusTooManyRequests, ""), true, true},
		{"429 with Retry-After at max", response(http.StatusTooManyRequests, "30"), true, true},
		{"429 with Retry-After above max", response(http.StatusTooManyRequests, "31"), false, false},
		{"503 with Retry-After", response(http.StatusServiceUnavailable, "0"), true, true},
		{"503 with Retry-After above max", response(http.StatusServiceUnavailable, "31"), false, false},
		{"503 without Retry-After", response(http.StatusServiceUnavailable, ""), true, false},
		{"503 with invalid Retry-After", response(http.StatusServiceUnavailable, "soon"), true, false},
		{"408", response(http.StatusRequestTimeout, ""), true, false},
		{"500", response(http.StatusInternalServerError, ""), true, false},
		{"502", response(http.StatusBadGateway, ""), true, false},
		{"504", response(http.StatusGatewayTimeout, ""), true, false},
		{"200", response(http.StatusOK, ""), false, false},
		{"401", response(http.StatusUnauthorized, ""), false, false},
		{"404", response(http.StatusNotFound, ""), false, false},
		{"501", response(http.StatusNotImplemented, ""), false, false},
		{"505", response(http.StatusHTTPVersionNotSupported, ""), false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for idempotent, want := range map[bool]bool{true: test.retryIdempotent, false: test.retryNonIdempotent} {
				retry, err := Default.checkRetry(contextWithState(idempotent), test.resp, nil)
				require.NoError(t, err)
				require.Equal(t, want, retry, "idempotent=%t", idempotent)
			}
		})
	}
}

func TestCheckRetry_Errors(t *testing.T) {
	connectionErr := &url.Error{Op: "Get", URL: "https://example.com", Err: &net.OpError{Op: "dial", Err: errors.New("connection refused")}}
	tlsAlert := &url.Error{Op: "Get", URL: "https://example.com", Err: &net.OpError{Op: "remote error", Err: errors.New("tls: certificate required")}}
	redirects := &url.Error{Op: "Get", URL: "https://example.com", Err: errors.New("stopped after 10 redirects")}

	tests := []struct {
		name               string
		err                error
		retryIdempotent    bool
		retryNonIdempotent bool
	}{
		{"connection error", connectionErr, true, false},
		{"TLS alert", tlsAlert, false, false},
		{"too many redirects", redirects, false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for idempotent, want := range map[bool]bool{true: test.retryIdempotent, false: test.retryNonIdempotent} {
				retry, err := Default.checkRetry(contextWithState(idempotent), nil, test.err)
				require.NoError(t, err)
				require.Equal(t, want, retry, "idempotent=%t", idempotent)
			}
		})
	}
}

func TestCheckRetry_FinishedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(contextWithState(true))
	cancel()

	retry, err := Default.checkRetry(ctx, response(http.StatusServiceUnavailable, ""), nil)
	require.NoError(t, err)
	require.False(t, retry)

	retry, err = Default.checkRetry(ctx, nil, &url.Error{Op: "Get", URL: "https://example.com", Err: context.Canceled})
	require.NoError(t, err)
	require.False(t, retry)
}

func TestParseRetryAfter(t *testing.T) {
	future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
	past := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)

	tests := []struct {
		name   string
		resp   *http.Response
		wantOk bool
		check  func(*testing.T, time.Duration)
	}{
		{"seconds", response(http.StatusTooManyRequests, "7"), true, func(t *testing.T, d time.Duration) { require.Equal(t, 7*time.Second, d) }},
		{"zero", response(http.StatusTooManyRequests, "0"), true, func(t *testing.T, d time.Duration) { require.Zero(t, d) }},
		{"future date", response(http.StatusServiceUnavailable, future), true, func(t *testing.T, d time.Duration) { require.InDelta(t, 10*time.Second, d, float64(2*time.Second)) }},
		{"past date", response(http.StatusServiceUnavailable, past), true, func(t *testing.T, d time.Duration) { require.Zero(t, d) }},
		{"negative", response(http.StatusTooManyRequests, "-1"), false, nil},
		{"unparsable", response(http.StatusTooManyRequests, "soon"), false, nil},
		{"missing", response(http.StatusTooManyRequests, ""), false, nil},
		{"other status", response(http.StatusInternalServerError, "5"), false, nil},
		{"no response", nil, false, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, ok := parseRetryAfter(test.resp)
			require.Equal(t, test.wantOk, ok)
			if test.check != nil {
				test.check(t, d)
			}
		})
	}
}
