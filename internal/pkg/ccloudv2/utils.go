package ccloudv2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	"github.com/hashicorp/go-retryablehttp"

	plog "github.com/confluentinc/cli/internal/pkg/log"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const (
	pageTokenQueryParameter = "page_token"
	ccloudV2ListPageSize    = 100
)

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

func newRetryableHttpClient() *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = new(plog.LeveledLogger)
	return client.StandardClient()
}

func getServerUrl(baseURL string, isTest bool) string {
	if isTest {
		return testserver.TestV2CloudURL.String()
	}
	if strings.Contains(baseURL, "devel") {
		return "https://api.devel.cpdev.cloud"
	} else if strings.Contains(baseURL, "stag") {
		return "https://api.stag.cpdev.cloud"
	}
	return "https://api.confluent.cloud"
}

func extractCmkNextPagePageToken(nextPageUrlStringNullable cmkv2.NullableString) (string, bool, error) {
	if nextPageUrlStringNullable.IsSet() {
		nextPageUrlString := *nextPageUrlStringNullable.Get()
		pageToken, err := extractPageToken(nextPageUrlString)
		return pageToken, false, err
	} else {
		return "", true, nil
	}
}

func extractIamNextPagePageToken(nextPageUrlStringNullable iamv2.NullableString) (string, bool, error) {
	if nextPageUrlStringNullable.IsSet() {
		nextPageUrlString := *nextPageUrlStringNullable.Get()
		pageToken, err := extractPageToken(nextPageUrlString)
		return pageToken, false, err
	} else {
		return "", true, nil
	}
}

func extractOrgNextPagePageToken(nextPageUrlStringNullable orgv2.NullableString) (string, bool, error) {
	if nextPageUrlStringNullable.IsSet() {
		nextPageUrlString := *nextPageUrlStringNullable.Get()
		pageToken, err := extractPageToken(nextPageUrlString)
		return pageToken, false, err
	} else {
		return "", true, nil
	}
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
