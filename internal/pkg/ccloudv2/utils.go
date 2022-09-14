package ccloudv2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	plog "github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const (
	pageTokenQueryParameter = "page_token"
	ccloudV2ListPageSize    = 100
)

var retryableHTTPCodes = map[int]bool{
	// 429
	http.StatusTooManyRequests: true,

	// 5XX
	http.StatusInternalServerError:           true,
	http.StatusNotImplemented:                true,
	http.StatusBadGateway:                    true,
	http.StatusServiceUnavailable:            true,
	http.StatusGatewayTimeout:                true,
	http.StatusHTTPVersionNotSupported:       true,
	http.StatusVariantAlsoNegotiates:         true,
	http.StatusInsufficientStorage:           true,
	http.StatusLoopDetected:                  true,
	http.StatusNotExtended:                   true,
	http.StatusNetworkAuthenticationRequired: true,
}

var Hostnames = []string{"confluent.cloud", "cpdev.cloud"}

func IsCCloudURL(url string, isTest bool) bool {
	for _, hostname := range Hostnames {
		if strings.Contains(url, hostname) {
			return true
		}
	}
	if isTest {
		return strings.Contains(url, testserver.TestCloudURL.Host) || strings.Contains(url, testserver.TestV2CloudURL.Host)
	}
	return false
}

func newRetryableHttpClient(unsafeTrace bool) *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = plog.NewLeveledLogger(unsafeTrace)
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ok, _ := retryableHTTPCodes[resp.StatusCode]; ok {
			return true, nil
		}
		return false, nil
	}
	return client.StandardClient()
}

func getServerUrl(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	if utils.Contains([]string{"confluent.cloud", "devel.cpdev.cloud", "stag.cpdev.cloud"}, u.Host) {
		u.Host = "api." + u.Host
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
