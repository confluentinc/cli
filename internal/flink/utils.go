package flink

import (
	"errors"
	"io"
	"net/http"
)

// Returns the next page number and whether we need to fetch more pages or not.
func extractPageOptions(receivedItemsLength int, currentPageNumber int) (nextPageNumber int, done bool) {
	if receivedItemsLength == 0 {
		return currentPageNumber, true
	}
	return currentPageNumber + 1, false
}

// Creates a rich error message from the HTTP response and the SDK error if possible.
func parseSdkError(httpResp *http.Response, sdkErr error) error {
	// If there's an error, and the httpResp is populated, it may contain a more detailed error message.
	if sdkErr != nil && httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
			respBody, parseError := io.ReadAll(httpResp.Body)
			if parseError == nil {
				return errors.New(string(respBody))
			}
		}
	}
	// In case we can't parse the body, or if there's none, return the original error.
	return sdkErr
}
