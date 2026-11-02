package ccloudv2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

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
