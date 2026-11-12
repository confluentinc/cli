package ccloudv2

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

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
