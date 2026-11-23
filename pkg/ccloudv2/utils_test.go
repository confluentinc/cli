package ccloudv2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func TestIsCCloudURL_True(t *testing.T) {
	for _, url := range []string{
		"confluent.cloud",
		"https://confluent.cloud",
		"https://devel.cpdev.cloud/",
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
	} {
		isCCloud := IsCCloudURL(url, false)
		require.True(t, isCCloud, url+" should return true")
	}
}

func TestIsCCloudURL_False(t *testing.T) {
	for _, url := range []string{
		"example.com",
		"example.com:8090",
		"https://example.com",
	} {
		isCCloud := IsCCloudURL(url, false)
		require.False(t, isCCloud, url+" should return false")
	}
}

func TestGetServerUrl(t *testing.T) {
	m := map[string]string{
		"https://confluent.cloud":   "https://api.confluent.cloud",
		"https://devel.cpdev.cloud": "https://api.devel.cpdev.cloud",
		"https://stag.cpdev.cloud":  "https://api.stag.cpdev.cloud",
		"https://stag.cpdev.cloud/": "https://api.stag.cpdev.cloud",
	}

	for baseUrl, serverUrl := range m {
		assert.Equal(t, serverUrl, getServerUrl(baseUrl))
	}
}

func TestToLower(t *testing.T) {
	require.Equal(t, "sasl-ssl", ToLower("SASL_SSL"))
}

func TestToUpper(t *testing.T) {
	require.Equal(t, "SASL_SSL", ToUpper("sasl-ssl"))
}

func TestExtractPageToken(t *testing.T) {
	token, err := ExtractPageToken("https://example.com/results?page_token=20")
	require.NoError(t, err)
	require.Equal(t, "20", token)
}

func TestExtractPageToken_MissingToken(t *testing.T) {
	_, err := ExtractPageToken("https://example.com/results")
	require.ErrorContains(t, err, `could not parse the value for query parameter "page_token"`)
}

func TestRetryPolicy(t *testing.T) {
	get := &url.Error{Op: "Get", Err: syscall.ECONNRESET}
	post := &url.Error{Op: "Post", Err: syscall.ECONNRESET}
	refused := &url.Error{Op: "Get", Err: syscall.ECONNREFUSED}

	for _, tt := range []struct {
		name string
		resp *http.Response
		err  error
		want bool
	}{
		{"429", &http.Response{StatusCode: http.StatusTooManyRequests}, nil, true},
		{"500", &http.Response{StatusCode: http.StatusInternalServerError}, nil, true},
		{"501", &http.Response{StatusCode: http.StatusNotImplemented}, nil, false},
		{"404", &http.Response{StatusCode: http.StatusNotFound}, nil, false},
		{"200", &http.Response{StatusCode: http.StatusOK}, nil, false},
		{"connection error on GET", nil, get, true},
		{"connection error on POST", nil, post, false},
		{"connection refused on GET", nil, refused, false},
		{"non-url error", nil, errors.New("boom"), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := retryPolicy(context.Background(), tt.resp, tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRetryBackoff(t *testing.T) {
	for attempt := 0; attempt < 5; attempt++ {
		base := time.Second << attempt
		for range 100 {
			wait := retryBackoff(time.Second, 30*time.Second, attempt, nil)
			require.GreaterOrEqual(t, wait, time.Duration(float64(base)*0.8))
			require.LessOrEqual(t, wait, min(time.Duration(float64(base)*1.2), 30*time.Second))
		}
	}
}

func TestRetryBackoff_RetryAfter(t *testing.T) {
	resp := func(retryAfter string) *http.Response {
		return &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{retryAfter}}}
	}

	for range 100 {
		// Never earlier than the server asked, and at most 20% later
		wait := retryBackoff(time.Second, 30*time.Second, 0, resp("3"))
		require.GreaterOrEqual(t, wait, 3*time.Second)
		require.LessOrEqual(t, wait, time.Duration(3.6*float64(time.Second)))

		// A long Retry-After is capped
		wait = retryBackoff(time.Second, 30*time.Second, 0, resp("3600"))
		require.GreaterOrEqual(t, wait, maxRetryAfter)
		require.LessOrEqual(t, wait, time.Duration(1.2*float64(maxRetryAfter)))
	}
}

func TestRetryableClient_ExhaustedReturnsLastResponse(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewRetryableHttpClient(&config.Config{IsTest: true}, false)
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	require.Equal(t, retryMax+1, requests)
}
