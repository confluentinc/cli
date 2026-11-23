package flink

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func TestFormatDeletionRequestedTime(t *testing.T) {
	for timestamp, expected := range map[string]string{
		"2026-11-18T16:00:52.452Z": "2026-11-18 16:00:52.452 +0000 UTC",
		"not a timestamp":          "not a timestamp",
	} {
		metadata := &cmfsdk.EnvironmentMetadata{DeletionRequestedTimestamp: cmfsdk.PtrString(timestamp)}
		require.Equal(t, expected, formatDeletionRequestedTime(metadata))
	}
	require.Empty(t, formatDeletionRequestedTime(nil))
}
