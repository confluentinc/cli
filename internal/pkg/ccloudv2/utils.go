package ccloudv2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	plog "github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/types"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const (
	pageTokenQueryParameter = "page_token"
	ccloudV2ListPageSize    = 100
)

type NullableString interface {
	Get() *string
	IsSet() bool
}

var Hostnames = []string{"confluent.cloud", "cpdev.cloud"}

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

func newRetryableHttpClient(unsafeTrace bool) *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = plog.NewLeveledLogger(unsafeTrace)
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil {
			return false, err
		}
		return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500, err
	}
	return client.StandardClient()
}

func getServerUrl(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	if types.Contains([]string{"confluent.cloud", "devel.cpdev.cloud", "stag.cpdev.cloud"}, u.Host) {
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
