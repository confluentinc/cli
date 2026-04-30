package flink

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func TestExtractComputePoolPhase(t *testing.T) {
	newPool := func(status *map[string]any) cmfsdk.ComputePool {
		return cmfsdk.ComputePool{
			Metadata: cmfsdk.ComputePoolMetadata{Name: "test-pool"},
			Status:   status,
		}
	}

	tests := []struct {
		name   string
		status *map[string]any
		want   string
	}{
		{
			name:   "nil status",
			status: nil,
			want:   "",
		},
		{
			name:   "empty status map",
			status: &map[string]any{},
			want:   "",
		},
		{
			name:   "phase key missing",
			status: &map[string]any{"observedGeneration": int64(1)},
			want:   "",
		},
		{
			name:   "phase value nil",
			status: &map[string]any{"phase": nil},
			want:   "",
		},
		{
			name:   "phase is string",
			status: &map[string]any{"phase": "RUNNING"},
			want:   "RUNNING",
		},
		{
			name:   "phase is empty string",
			status: &map[string]any{"phase": ""},
			want:   "",
		},
		{
			name:   "phase is non-string (contract violation)",
			status: &map[string]any{"phase": 42},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractComputePoolPhase(newPool(tt.status))
			require.Equal(t, tt.want, got)
		})
	}
}
