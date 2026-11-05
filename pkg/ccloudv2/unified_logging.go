package ccloudv2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

// This file serves the unified `confluent logs` commands. It sits next to logging.go, which serves `confluent connect
// logs`, and reuses its logging host lookup without changing it.

// getLoggingServerUrl returns the logging API's server URL, without a path. The host is chosen by getLoggingUrl, so the
// mapping from a platform to its logging host stays in one place.
func getLoggingServerUrl(baseURL string) (string, error) {
	searchUrl, err := getLoggingUrl(baseURL)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(searchUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse logging URL: %w", err)
	}
	u.Path = ""
	return u.String(), nil
}

// newLoggingClient builds the SDK client for the logging API, which is served from its own host rather than the
// regular Confluent Cloud API URL, and a context carrying the data plane token it authenticates with.
func (c *Client) newLoggingClient() (*loggingv1.APIClient, context.Context, error) {
	loggingURL, err := getLoggingServerUrl(c.cfg.Context().GetPlatformServer())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get logging API URL: %w", err)
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get data plane token: %w", err)
	}

	cfg := loggingv1.NewConfiguration()
	cfg.HTTPClient = newRetryableHttpClientReturningErrors(c.cfg, false)
	cfg.Servers = loggingv1.ServerConfigurations{{URL: loggingURL}}
	cfg.UserAgent = c.cfg.Version.UserAgent

	ctx := context.WithValue(context.Background(), loggingv1.ContextAccessToken, dataplaneToken)
	return loggingv1.NewAPIClient(cfg), ctx, nil
}

// ListLogSources queries GET /logs/v1/sources.
func (c *Client) ListLogSources() (*loggingv1.LoggingV1SourceList, error) {
	client, ctx, err := c.newLoggingClient()
	if err != nil {
		return nil, err
	}

	sources, httpResp, err := client.LogsLoggingV1Api.ListLoggingV1Sources(ctx).Execute()
	if err != nil {
		return nil, catchLoggingError(err, httpResp)
	}
	return &sources, nil
}

// catchLoggingError turns a logging API failure into a CLI error. A 429 tells the user how long to wait, from the
// Retry-After header the API sends.
func catchLoggingError(err error, httpResp *http.Response) error {
	if httpResp != nil && httpResp.StatusCode == http.StatusTooManyRequests {
		return rateLimitError(httpResp.Header.Get("Retry-After"), time.Now())
	}
	return errors.CatchCCloudV2Error(err, httpResp)
}

// rateLimitError reports a 429. Retry-After is either a number of seconds or an HTTP date.
func rateLimitError(retryAfter string, now time.Time) error {
	const message = "rate limit exceeded for log queries"

	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds > 0 {
		return fmt.Errorf("%s. Retry after %ds", message, seconds)
	}
	if date, err := http.ParseTime(retryAfter); err == nil {
		if wait := date.Sub(now); wait > 0 {
			return fmt.Errorf("%s. Retry after %ds", message, int((wait+time.Second-1)/time.Second))
		}
	}
	return fmt.Errorf("%s. Try again later", message)
}
