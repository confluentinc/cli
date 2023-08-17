package flink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatUtcOffsetToTimezone(t *testing.T) {
	testCases := []struct {
		offsetSeconds int
		expected      string
	}{
		{
			offsetSeconds: hoursToSeconds(5.5),
			expected:      "UTC+05:30",
		},
		{
			offsetSeconds: hoursToSeconds(-6),
			expected:      "UTC-06:00",
		},
		{
			offsetSeconds: hoursToSeconds(0),
			expected:      "UTC+00:00",
		},
		{
			offsetSeconds: hoursToSeconds(-2.25),
			expected:      "UTC-02:15",
		},
		{
			offsetSeconds: hoursToSeconds(3.75),
			expected:      "UTC+03:45",
		},
	}

	for _, tc := range testCases {
		actual := formatUtcOffsetToTimezone(tc.offsetSeconds)
		require.Equal(t, tc.expected, actual)
	}
}

func hoursToSeconds(hours float32) int {
	return int(hours * 60 * 60)
}
