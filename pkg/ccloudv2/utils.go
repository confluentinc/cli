package ccloudv2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/log"
	plog "github.com/confluentinc/cli/v4/pkg/log"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const (
	pageTokenQueryParameter = "page_token"
	ccloudV2ListPageSize    = 100
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
	client := retryablehttp.NewClient()
	client.Logger = plog.NewLeveledLogger(unsafeTrace)
	client.CheckRetry = func(_ context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil {
			return false, err
		}

		if resp.StatusCode == http.StatusUnauthorized && cfg.Context().GetState().IsExpired() {
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

		return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500, err
	}

	return client.StandardClient()
}

func NewRetryableHttpClientWithRedirect(unsafeTrace bool, checkRedirect func(*http.Request, []*http.Request) error) *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = plog.NewLeveledLogger(unsafeTrace)
	client.CheckRetry = func(_ context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil {
			return false, err
		}
		return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500, err
	}
	client.HTTPClient.CheckRedirect = checkRedirect
	return client.StandardClient()
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

	if slices.Contains([]string{"confluent.cloud", "devel.cpdev.cloud", "stag.cpdev.cloud"}, u.Host) {
		u.Host = "api." + u.Host
		u.Path = ""
	} else {
		u.Path = "api"
	}

	return u.String()
}

func extractPageToken(nextPageUrlString string) (string, error) {
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
	if !nextPageUrl.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrl.Get()
	if nextPageUrlString == "" {
		return "", true, nil
	}
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
