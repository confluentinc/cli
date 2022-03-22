package ccloudv2

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

const (
	pageTokenQueryParameter = "page_token"
)

func getServerUrl(baseURL string, isTest bool) string {
	if isTest {
		return "http://127.0.0.1:2048"
	}
	if strings.Contains(baseURL, "devel") {
		return "https://api.devel.cpdev.cloud"
	} else if strings.Contains(baseURL, "stag") {
		return "https://api.stag.cpdev.cloud"
	}
	return "https://api.confluent.cloud"
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

func extractPageToken(nextPageUrlString string) (string, error) {
	nextPageUrl, err := url.Parse(nextPageUrlString)
	if err != nil {
		log.Printf("[ERROR] Could not parse %s into URL, %s", nextPageUrlString, err)
		return "", err
	}
	pageToken := nextPageUrl.Query().Get(pageTokenQueryParameter)
	if pageToken == "" {
		return "", fmt.Errorf("[ERROR] Could not parse the value for %s query parameter from %s", pageTokenQueryParameter, nextPageUrlString)
	}
	return pageToken, nil
}
