package ccloudv2

import (
	"fmt"
	"log"
	"net/url"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

func extractNextPagePageToken(nextPageUrlStringNullable iamv2.NullableString) (string, bool, error) {
	var err error
	var pageToken string
	if nextPageUrlStringNullable.IsSet() {
		nextPageUrlString := *nextPageUrlStringNullable.Get()
		pageToken, err = extractPageToken(nextPageUrlString)
	} else {
		return pageToken, true, nil
	}
	return pageToken, false, err
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
