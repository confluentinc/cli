package ccloudv2

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/httpclient"
)

// requireRateLimitError checks a 429 error's message and suggestion.
func requireRateLimitError(t *testing.T, err error, suggestion string) {
	t.Helper()
	require.EqualError(t, err, logsRateLimitErrorMsg)
	var withSuggestions errors.ErrorWithSuggestions
	require.ErrorAs(t, err, &withSuggestions)
	require.Equal(t, suggestion, withSuggestions.GetSuggestionsMsg())
}

func TestCatchLoggingErrorRateLimit(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"12"}}}
	requireRateLimitError(t, catchLoggingError(nil, resp), "Retry after 12s.")

	// Without the header there is nothing to promise.
	resp = &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	requireRateLimitError(t, catchLoggingError(nil, resp), "Try again later.")
}

func TestRateLimitError(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	requireRateLimitError(t, rateLimitError("30", now), "Retry after 30s.")
	requireRateLimitError(t, rateLimitError(now.Add(90*time.Second).Format(http.TimeFormat), now), "Retry after 90s.")
	requireRateLimitError(t, rateLimitError(now.Add(-time.Minute).Format(http.TimeFormat), now), "Try again later.")
	requireRateLimitError(t, rateLimitError("0", now), "Try again later.")
	requireRateLimitError(t, rateLimitError("soon", now), "Try again later.")
}

func TestDecodeLogsSearchResponse_PreservesLargeIntegers(t *testing.T) {
	// 9007199254740993 is 2^53 + 1, the first integer a float64 cannot represent.
	body := []byte(`{"data":[{"id":"evt-1","offset":9007199254740993,"ratio":0.5,"nested":{"ids":[18446744073709551615,-9007199254740993]}}],"api_version":"v1","kind":"Logs"}`)

	response, err := decodeLogsSearchResponse(body)
	require.NoError(t, err)
	require.Len(t, response.Data, 1)

	jsonOut, err := json.Marshal(response.Data)
	require.NoError(t, err)
	require.Contains(t, string(jsonOut), `"offset":9007199254740993`)
	require.Contains(t, string(jsonOut), `"ratio":0.5`)
	require.Contains(t, string(jsonOut), `"ids":[18446744073709551615,-9007199254740993]`)

	yamlOut, err := yaml.Marshal(response.Data)
	require.NoError(t, err)
	require.Contains(t, string(yamlOut), "offset: 9007199254740993")
	require.Contains(t, string(yamlOut), "ratio: 0.5")

	// JSON lines are one json.Marshal per record.
	line, err := json.Marshal(response.Data[0])
	require.NoError(t, err)
	require.Contains(t, string(line), `"offset":9007199254740993`)
}

func TestDecodeLogsSearchResponse_InvalidBody(t *testing.T) {
	_, err := decodeLogsSearchResponse([]byte(`not json`))
	require.ErrorContains(t, err, "failed to decode logging API response")
}

// newTestLoggingHttpClient is the logging client with its retry waits shortened, so tests do not sleep.
func newTestLoggingHttpClient() *http.Client {
	cfg := newLoggingHttpConfig()
	cfg.RetryWaitMin = time.Millisecond
	cfg.RetryWaitMax = 2 * time.Millisecond
	return httpclient.New(cfg)
}

func TestLoggingHttpClientReturnsRateLimitAtOnce(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Retry-After", "12")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	resp, err := newTestLoggingHttpClient().Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	require.Equal(t, "12", resp.Header.Get("Retry-After"))
	require.EqualValues(t, 1, requests.Load())
}

func TestLoggingHttpClientRetriesServerErrorThenReturnsLastResponse(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("backend unavailable"))
	}))
	defer server.Close()

	resp, err := newTestLoggingHttpClient().Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "backend unavailable", string(body))
	require.Greater(t, requests.Load(), int32(1))
}

func TestLoggingHttpClientRetriesServerErrorUntilSuccess(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := newTestLoggingHttpClient().Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.EqualValues(t, 3, requests.Load())
}

func TestLoggingHttpClientDoesNotRetryClientError(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	resp, err := newTestLoggingHttpClient().Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.EqualValues(t, 1, requests.Load())
}
