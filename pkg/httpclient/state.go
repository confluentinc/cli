package httpclient

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

type stateKey struct{}

type idempotentKey struct{}

// state is per request, stored in the request's context. Hooks for one request run in sequence.
type state struct {
	idempotent bool // set by stateTransport
	attempt    int  // set by recordAttempt; 0 is the first attempt
}

func stateFrom(ctx context.Context) *state {
	st, _ := ctx.Value(stateKey{}).(*state)
	return st
}

// MarkReadOnly marks a POST that changes nothing (such as a Metrics query), so it is retried like a GET.
func MarkReadOnly(ctx context.Context) context.Context {
	return context.WithValue(ctx, idempotentKey{}, true)
}

func isIdempotent(req *http.Request) bool {
	switch req.Method {
	case "", http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPut, http.MethodDelete:
		return true
	}
	idempotent, _ := req.Context().Value(idempotentKey{}).(bool)
	return idempotent
}

// recordAttempt is a retryablehttp.RequestLogHook, used only to count attempts
func recordAttempt(_ retryablehttp.Logger, req *http.Request, attempt int) {
	if st := stateFrom(req.Context()); st != nil {
		st.attempt = attempt
	}
}

type stateTransport struct {
	next http.RoundTripper
}

func (t *stateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	st := &state{idempotent: isIdempotent(req)}

	resp, err := t.next.RoundTrip(req.WithContext(context.WithValue(req.Context(), stateKey{}, st)))
	if err == nil {
		return resp, nil
	}

	// A RoundTripper must not return both; the library only does so with an already-consumed response.
	if resp != nil {
		_ = resp.Body.Close()
	}

	return nil, st.wrap(err)
}

// wrap returns err unchanged after one attempt that can't have been applied; otherwise a RetryError
func (st *state) wrap(err error) error {
	attempts := st.attempt + 1
	maybeApplied := !st.idempotent && !notSent(err)
	// No retries and nothing to warn about, so a RetryError adds nothing
	if attempts == 1 && !maybeApplied {
		return err
	}

	return &RetryError{Attempts: attempts, Err: err, MaybeApplied: maybeApplied}
}

// notSent reports whether the request failed before reaching the server (DNS, connect, proxy, TLS handshake).
func notSent(err error) bool {
	// DNS failures arrive as a dial error.
	if opErr, ok := errors.AsType[*net.OpError](err); ok && (opErr.Op == "dial" || opErr.Op == "proxyconnect" || opErr.Op == "remote error") {
		return true
	}

	if _, ok := errors.AsType[*tls.CertificateVerificationError](err); ok {
		return true
	}

	if _, ok := errors.AsType[tls.RecordHeaderError](err); ok {
		return true
	}

	// net/http's handshake timeout error type is unexported.
	return strings.Contains(err.Error(), "net/http: TLS handshake timeout")
}
