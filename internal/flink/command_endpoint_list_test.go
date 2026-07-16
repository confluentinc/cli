package flink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlinkEndpointUrl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bare FQDN gets https:// prefix",
			input:    "flink.us-east-1.aws.confluent.cloud",
			expected: "https://flink.us-east-1.aws.confluent.cloud",
		},
		{
			name:     "multi-PLATT GLB access-point FQDN gets https:// prefix",
			input:    "flink-ap4l9zl0.ap-northeast-1.aws.accesspoint.glb.confluent.cloud",
			expected: "https://flink-ap4l9zl0.ap-northeast-1.aws.accesspoint.glb.confluent.cloud",
		},
		{
			name:     "value with https:// scheme is returned unchanged",
			input:    "https://flink.eu-west-1.aws.confluent.cloud",
			expected: "https://flink.eu-west-1.aws.confluent.cloud",
		},
		{
			name:     "value with http:// scheme is returned unchanged",
			input:    "http://127.0.0.1:1026",
			expected: "http://127.0.0.1:1026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, flinkEndpointUrl(tt.input))
		})
	}
}
