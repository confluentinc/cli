package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsIdempotent(t *testing.T) {
	for method, want := range map[string]bool{
		"":                 true,
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodOptions: true,
		http.MethodTrace:   true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodPost:    false,
		http.MethodPatch:   false,
	} {
		req, err := http.NewRequest(method, "https://example.com", nil)
		require.NoError(t, err)
		require.Equal(t, want, isIdempotent(req), method)
	}
}

func TestIsIdempotent_MarkReadOnly(t *testing.T) {
	req, err := http.NewRequestWithContext(MarkReadOnly(context.Background()), http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	require.True(t, isIdempotent(req))
}

func TestWrap(t *testing.T) {
	sent := &net.OpError{Op: "read", Err: errors.New("connection reset by peer")}
	dial := &net.OpError{Op: "dial", Err: errors.New("connection refused")}

	tests := []struct {
		name      string
		st        state
		err       error
		wantRetry *RetryError
	}{
		{"one attempt, idempotent", state{idempotent: true}, sent, nil},
		{"one attempt, not sent", state{}, dial, nil},
		{"several attempts", state{idempotent: true, attempt: 4}, sent, &RetryError{Attempts: 5, Err: sent}},
		{"maybe applied", state{}, sent, &RetryError{Attempts: 1, Err: sent, MaybeApplied: true}},
		{"maybe applied after retries", state{attempt: 1}, sent, &RetryError{Attempts: 2, Err: sent, MaybeApplied: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.st.wrap(test.err)
			if test.wantRetry == nil {
				require.Same(t, test.err, got)
			} else {
				require.Equal(t, test.wantRetry, got)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

func TestStateTransport_ResponseWithError(t *testing.T) {
	// The library returns both when it can't rewind the request body.
	body := &trackedBody{Reader: strings.NewReader("")}
	cause := errors.New("failed to rewind body")
	client := &http.Client{Transport: &stateTransport{next: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: body}, cause
	})}}

	var stdLog bytes.Buffer
	log.SetOutput(&stdLog)
	defer log.SetOutput(os.Stderr)

	resp, err := client.Get("https://example.com")
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Nil(t, resp)
	require.ErrorIs(t, err, cause)
	require.True(t, body.closed)
	require.Empty(t, stdLog.String(), "stateTransport returned a response with its error, so net/http logged to stderr")
}

func TestNotSent(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"dial", &net.OpError{Op: "dial", Err: errors.New("connection refused")}, true},
		{"proxy", &net.OpError{Op: "proxyconnect", Err: errors.New("connection refused")}, true},
		{"TLS alert", &net.OpError{Op: "remote error", Err: errors.New("tls: bad certificate")}, true},
		{"certificate", &tls.CertificateVerificationError{Err: errors.New("unknown authority")}, true},
		{"record header", tls.RecordHeaderError{Msg: "first record does not look like a TLS handshake"}, true},
		{"handshake timeout", errors.New("net/http: TLS handshake timeout"), true},
		{"wrapped dial", &url.Error{Op: "Post", URL: "https://example.com", Err: &net.OpError{Op: "dial", Err: errors.New("refused")}}, true},
		{"read", &net.OpError{Op: "read", Err: errors.New("connection reset by peer")}, false},
		{"timeout awaiting headers", errors.New("context deadline exceeded (Client.Timeout exceeded while awaiting headers)"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, notSent(test.err))
		})
	}
}
