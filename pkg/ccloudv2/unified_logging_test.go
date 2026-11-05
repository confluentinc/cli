package ccloudv2

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCatchLoggingErrorRateLimit(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"12"}}}
	require.EqualError(t, catchLoggingError(nil, resp), "rate limit exceeded for log queries. Retry after 12s")

	// Without the header there is nothing to promise.
	resp = &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	require.EqualError(t, catchLoggingError(nil, resp), "rate limit exceeded for log queries. Try again later")
}

func TestRateLimitError(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	require.EqualError(t, rateLimitError("30", now), "rate limit exceeded for log queries. Retry after 30s")
	require.EqualError(t, rateLimitError(now.Add(90*time.Second).Format(http.TimeFormat), now), "rate limit exceeded for log queries. Retry after 90s")
	require.EqualError(t, rateLimitError(now.Add(-time.Minute).Format(http.TimeFormat), now), "rate limit exceeded for log queries. Try again later")
	require.EqualError(t, rateLimitError("0", now), "rate limit exceeded for log queries. Try again later")
	require.EqualError(t, rateLimitError("soon", now), "rate limit exceeded for log queries. Try again later")
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
