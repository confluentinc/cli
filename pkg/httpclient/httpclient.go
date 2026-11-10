// Package httpclient builds the CLI's HTTP clients on go-retryablehttp: per-attempt timeouts,
// intentional retry limits, and jittered exponential backoff that honors Retry-After.
package httpclient

import (
	"crypto/tls"
	"math/rand/v2"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const growth = 1.6

type Config struct {
	RetryMax     int           // retries after the first attempt
	RetryWaitMin time.Duration // first computed wait
	RetryWaitMax time.Duration // cap on computed waits; longest Retry-After honored
	Jitter       float64       // plus/minus jitter on computed waits, plus 0 to jitter on Retry-After

	AttemptTimeout        time.Duration // one attempt, including reading the body
	ConnectTimeout        time.Duration // DNS and TCP connect
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration // zero means only AttemptTimeout applies

	TLSConfig     *tls.Config                                // custom CA and/or client certificate
	CheckRedirect func(*http.Request, []*http.Request) error // nil follows up to 10 redirects

	// Tests swap these out to control jitter and capture notices; New defaults them to rand.Float64 and stderr
	random      func() float64
	printNotice func(string)
}

var (
	// Default is used by all commands, including the Flink shell and `ai` sessions.
	Default = Config{
		RetryMax:            4,
		RetryWaitMin:        1 * time.Second,
		RetryWaitMax:        30 * time.Second,
		Jitter:              0.2,
		AttemptTimeout:      60 * time.Second,
		ConnectTimeout:      10 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// NoRetry is for background calls the user doesn't see: usage report, feedback, feature flags, update check.
	NoRetry = Config{
		RetryMax:            0,
		AttemptTimeout:      5 * time.Second,
		ConnectTimeout:      10 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// LargeDownload is for Hub plugin archives and the CLI binary. A retry restarts the download.
	LargeDownload = Config{
		RetryMax:              4,
		RetryWaitMin:          1 * time.Second,
		RetryWaitMax:          30 * time.Second,
		Jitter:                0.2,
		AttemptTimeout:        30 * time.Minute,
		ConnectTimeout:        10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
)

func New(cfg Config) *http.Client {
	if cfg.random == nil {
		cfg.random = rand.Float64
	}
	if cfg.printNotice == nil {
		cfg.printNotice = func(s string) { output.ErrPrintln(false, s) }
	}

	client := &retryablehttp.Client{
		HTTPClient: &http.Client{
			Transport:     cfg.transport(),
			Timeout:       cfg.AttemptTimeout,
			CheckRedirect: cfg.CheckRedirect,
		},
		Logger:         nil, // TODO: plog.LeveledLogger once its key/value formatting is fixed
		RetryWaitMin:   cfg.RetryWaitMin,
		RetryWaitMax:   cfg.RetryWaitMax,
		RetryMax:       cfg.RetryMax,
		RequestLogHook: recordAttempt,
		CheckRetry:     cfg.checkRetry,
		Backoff:        cfg.backoff,
		ErrorHandler:   retryablehttp.PassthroughErrorHandler,
	}

	return &http.Client{
		Transport: &stateTransport{next: &retryablehttp.RoundTripper{Client: client}},
		// Only the inner client follows redirects.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func (cfg Config) transport() *http.Transport {
	transport := utils.DefaultTransport()
	// KeepAlive matches the CLI's existing transports (and net/http's) instead of net.Dialer's 15s default.
	transport.DialContext = (&net.Dialer{Timeout: cfg.ConnectTimeout, KeepAlive: 30 * time.Second}).DialContext
	transport.TLSHandshakeTimeout = cfg.TLSHandshakeTimeout
	transport.ResponseHeaderTimeout = cfg.ResponseHeaderTimeout

	if cfg.TLSConfig != nil {
		tlsConfig := cfg.TLSConfig.Clone()
		tlsConfig.MinVersion = max(tlsConfig.MinVersion, tls.VersionTLS12)
		transport.TLSClientConfig = tlsConfig
	}

	return transport
}
