package ccloudv2

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/log"
	plog "github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const (
	pageTokenQueryParameter = "page_token"
	ccloudV2ListPageSize    = 100

	retryMax       = 5
	requestTimeout = 30 * time.Second
	maxRetryAfter  = time.Minute
	retryJitter    = 0.2
	longRetryWait  = 5 * time.Second
)

var (
	ByocSupportClouds = []string{"aws", "azure", "gcp"}
)

var (
	ProviderIntegrationSupportClouds = []string{"aws"}
)

type NullableString interface {
	Get() *string
	IsSet() bool
}

var Hostnames = []string{
	"confluent.cloud",
	"confluentgov-internal.com",
	"confluentgov.com",
	"cpdev.cloud",
}

func IsCCloudURL(url string, isTest bool) bool {
	for _, hostname := range Hostnames {
		if strings.Contains(url, hostname) {
			return true
		}
	}
	if isTest {
		return strings.Contains(url, testserver.TestCloudUrl.Host) || strings.Contains(url, testserver.TestV2CloudUrl.Host)
	}
	return false
}

func NewRetryableHttpClient(cfg *config.Config, unsafeTrace bool) *http.Client {
	client := newRetryableClient(cfg, unsafeTrace)
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized && cfg.Context().GetState().IsExpired() {
			params := &ccloudv1.Params{
				BaseURL:    cfg.Context().GetPlatformServer(),
				HttpClient: ccloudv1.BaseClient,
				Logger:     log.CliLogger,
				UserAgent:  cfg.Version.UserAgent,
			}
			v1Client := ccloudv1.NewClient(params)

			_ = cfg.Context().RefreshSession(v1Client)
			_ = cfg.Save()

			return true, err
		}

		return retryPolicy(ctx, resp, err)
	}

	return client.StandardClient()
}

func NewRetryableHttpClientWithRedirect(unsafeTrace bool, checkRedirect func(*http.Request, []*http.Request) error) *http.Client {
	client := newRetryableClient(nil, unsafeTrace)
	client.CheckRetry = retryPolicy
	client.HTTPClient.CheckRedirect = checkRedirect
	return client.StandardClient()
}

func newRetryableClient(cfg *config.Config, unsafeTrace bool) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = plog.NewLeveledLogger(unsafeTrace)
	client.RetryMax = retryMax
	client.HTTPClient.Timeout = requestTimeout
	client.Backoff = retryBackoff
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler

	// Don't sleep between retries in integration tests
	if cfg != nil && cfg.IsTest {
		client.RetryWaitMin, client.RetryWaitMax = 0, 0
	}

	return client
}

// Use the library's retry policy (429, 5xx except 501, connection errors), except that a connection error is only
// retried for idempotent methods, since the server may have already processed the request, and never when the
// connection was refused, since nothing is listening.
func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil && (!isIdempotent(err) || errors.Is(err, syscall.ECONNREFUSED)) {
		return false, err
	}
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

func isIdempotent(err error) bool {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return false
	}
	return slices.Contains([]string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions}, strings.ToUpper(urlErr.Op))
}

// Exponential backoff with jitter. A Retry-After wait is capped, and jitter only extends it, so we never retry
// earlier than the server asked.
func retryBackoff(minWait, maxWait time.Duration, attempt int, resp *http.Response) time.Duration {
	wait := retryablehttp.DefaultBackoff(minWait, maxWait, attempt, resp)
	if resp != nil && resp.Header.Get("Retry-After") != "" {
		wait = min(wait, maxRetryAfter)
		wait += time.Duration(rand.Float64() * retryJitter * float64(wait))
	} else {
		wait = min(time.Duration(float64(wait)*(1+retryJitter*(2*rand.Float64()-1))), maxWait)
	}

	if wait > longRetryWait {
		output.ErrPrintf(false, "Waiting %s before retrying the request.\n", wait.Round(time.Second))
	}

	return wait
}

func ToLower(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", "-")
}

func ToUpper(s string) string {
	return strings.ReplaceAll(strings.ToUpper(s), "-", "_")
}

func getServerUrl(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	if slices.Contains([]string{"confluent.cloud", "devel.cpdev.cloud", "stag.cpdev.cloud", "dr-test.cpdev.cloud"}, u.Host) {
		u.Host = "api." + u.Host
		u.Path = ""
	} else {
		u.Path = "api"
	}

	return u.String()
}

// ExtractPageToken pulls page_token out of an absolute "next page" URL.
func ExtractPageToken(nextPageUrlString string) (string, error) {
	nextPageUrl, err := url.Parse(nextPageUrlString)
	if err != nil {
		plog.CliLogger.Errorf("Could not parse %s into URL, %v", nextPageUrlString, err)
		return "", err
	}
	pageToken := nextPageUrl.Query().Get(pageTokenQueryParameter)
	if pageToken == "" {
		return "", fmt.Errorf(`could not parse the value for query parameter "%s" from %s`, pageTokenQueryParameter, nextPageUrlString)
	}
	return pageToken, nil
}

func extractNextPageToken(nextPageUrl NullableString) (string, bool, error) {
	if !nextPageUrl.IsSet() || nextPageUrl.Get() == nil {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrl.Get()
	if nextPageUrlString == "" {
		return "", true, nil
	}
	pageToken, err := ExtractPageToken(nextPageUrlString)
	return pageToken, false, err
}
