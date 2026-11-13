package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func fastConfig(base Config) (Config, *atomic.Int32) {
	var backoffs atomic.Int32
	cfg := base
	cfg.RetryWaitMin = time.Millisecond
	cfg.RetryWaitMax = 5 * time.Millisecond
	cfg.random = func() float64 { backoffs.Add(1); return 0.5 }
	cfg.printNotice = func(string) {}
	return cfg, &backoffs
}

// statusServer responds with statuses in order, then repeats the last one.
func statusServer(t *testing.T, statuses ...int) (*httptest.Server, *atomic.Int32) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := int(attempts.Add(1))
		status := statuses[min(n, len(statuses))-1]
		w.WriteHeader(status)
		_, _ = w.Write([]byte(http.StatusText(status)))
	}))
	t.Cleanup(server.Close)
	return server, &attempts
}

func stallingServer(t *testing.T, stalls int32) (*httptest.Server, *atomic.Int32) {
	var attempts atomic.Int32
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The server notices a client hang-up only after the body is read.
		_, _ = io.Copy(io.Discard, r.Body)
		if attempts.Add(1) <= stalls {
			select {
			case <-r.Context().Done():
			case <-done:
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(done) })
	return server, &attempts
}

func tlsServer(t *testing.T, clientAuth tls.ClientAuthType) (*httptest.Server, *atomic.Int32) {
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.TLS = &tls.Config{ClientAuth: clientAuth}
	server.StartTLS()
	t.Cleanup(server.Close)
	return server, &connections
}

func closedServerURL() string {
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	return server.URL
}

func do(t *testing.T, client *http.Client, method, url, body string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	require.NoError(t, err)
	return client.Do(req)
}

func readBody(t *testing.T, resp *http.Response) string {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestNew_RetriesUntilSuccess(t *testing.T) {
	server, attempts := statusServer(t, http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusOK)
	cfg, _ := fastConfig(Default)

	resp, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "OK", readBody(t, resp))
	require.Equal(t, int32(3), attempts.Load())
}

func TestNew_ReturnsLastResponseWhenRetriesRunOut(t *testing.T) {
	server, attempts := statusServer(t, http.StatusServiceUnavailable)
	cfg, backoffs := fastConfig(Default)

	resp, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, "Service Unavailable", readBody(t, resp))
	require.Equal(t, int32(cfg.RetryMax+1), attempts.Load())
	require.Equal(t, int32(cfg.RetryMax), backoffs.Load(), "backoff ran a different number of times than there were retries")
}

func TestNew_PostNotRetriedOnServerError(t *testing.T) {
	server, attempts := statusServer(t, http.StatusInternalServerError, http.StatusOK)
	cfg, _ := fastConfig(Default)

	resp, err := do(t, New(cfg), http.MethodPost, server.URL, `{"name":"example"}`)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	_ = readBody(t, resp)
	require.Equal(t, int32(1), attempts.Load())
}

func TestNew_PostRetriedOnTooManyRequests(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	// Records each body to check that the retried POST resends it intact
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		defer mu.Unlock()
		bodies = append(bodies, string(body))
		if len(bodies) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	cfg, _ := fastConfig(Default)

	resp, err := do(t, New(cfg), http.MethodPost, server.URL, `{"name":"example"}`)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = readBody(t, resp)
	require.Equal(t, []string{`{"name":"example"}`, `{"name":"example"}`}, bodies)
}

func TestNew_NoRetry(t *testing.T) {
	server, attempts := statusServer(t, http.StatusServiceUnavailable)
	cfg, _ := fastConfig(NoRetry)

	resp, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	_ = readBody(t, resp)
	require.Equal(t, int32(1), attempts.Load())
}

func TestNew_ConnectionErrors(t *testing.T) {
	url := closedServerURL()
	cfg, _ := fastConfig(Default)
	client := New(cfg)

	_, err := do(t, client, http.MethodGet, url, "")
	var retryErr *RetryError
	require.ErrorAs(t, err, &retryErr)
	require.Equal(t, cfg.RetryMax+1, retryErr.Attempts)
	require.False(t, retryErr.MaybeApplied)

	_, err = do(t, client, http.MethodPost, url, "{}")
	require.Error(t, err)
	require.False(t, errors.As(err, &retryErr), "a POST that never connected was wrapped in a RetryError instead of keeping its original error")
	require.Contains(t, err.Error(), "connection refused")
}

func TestNew_AttemptTimeout(t *testing.T) {
	server, attempts := stallingServer(t, 1)
	cfg, _ := fastConfig(Default)
	cfg.AttemptTimeout = 100 * time.Millisecond

	resp, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = readBody(t, resp)
	require.Equal(t, int32(2), attempts.Load())
}

func TestNew_AttemptTimeoutOnPost(t *testing.T) {
	server, attempts := stallingServer(t, 1)
	cfg, _ := fastConfig(Default)
	cfg.AttemptTimeout = 100 * time.Millisecond

	_, err := do(t, New(cfg), http.MethodPost, server.URL, "{}")
	var retryErr *RetryError
	require.ErrorAs(t, err, &retryErr)
	require.True(t, retryErr.MaybeApplied)
	require.Equal(t, 1, retryErr.Attempts)
	require.Equal(t, int32(1), attempts.Load())
}

func TestNew_ContextCanceledDuringWait(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.Header().Set("Retry-After", "20")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := Default
	cfg.printNotice = func(string) {}
	cfg.random = func() float64 { cancel(); return 0.5 } // ends the request just as the wait begins

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	start := time.Now()
	_, err = New(cfg).Do(req)
	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, time.Since(start), 5*time.Second)
	require.Equal(t, int32(1), attempts.Load())
}

func TestNew_Redirects(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.RedirectHandler(target.URL, http.StatusFound))
	defer redirect.Close()

	resp, err := do(t, New(Default), http.MethodGet, redirect.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = readBody(t, resp)

	cfg := Default
	cfg.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err = do(t, New(cfg), http.MethodGet, redirect.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, resp.StatusCode)
	_ = readBody(t, resp)
}

func TestNew_UntrustedServerCertificate(t *testing.T) {
	server, connections := tlsServer(t, tls.NoClientCert)
	cfg, _ := fastConfig(Default)

	_, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	var certErr *tls.CertificateVerificationError
	require.ErrorAs(t, err, &certErr)
	require.Equal(t, int32(1), connections.Load())
}

func TestNew_RejectedClientCertificate(t *testing.T) {
	server, connections := tlsServer(t, tls.RequireAndVerifyClientCert)
	cfg, _ := fastConfig(Default)
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	cfg.TLSConfig = &tls.Config{RootCAs: roots}

	_, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	require.Error(t, err)
	require.True(t, isTLSAlert(err), err.Error())
	require.Equal(t, int32(1), connections.Load())
}

func TestNew_TLSConfig(t *testing.T) {
	server, _ := tlsServer(t, tls.NoClientCert)
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	cfg := Default
	cfg.TLSConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS10}

	resp, err := do(t, New(cfg), http.MethodGet, server.URL, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = readBody(t, resp)
	require.Equal(t, uint16(tls.VersionTLS12), cfg.transport().TLSClientConfig.MinVersion)
	require.Equal(t, uint16(tls.VersionTLS10), cfg.TLSConfig.MinVersion, "New modified the caller's TLSConfig")
}
