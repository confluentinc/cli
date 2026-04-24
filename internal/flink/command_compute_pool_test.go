package flink

import (
	"testing"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func TestExtractComputePoolPhase(t *testing.T) {
	tests := []struct {
		name   string
		status *map[string]interface{}
		want   string
	}{
		{
			name:   "nil status pointer",
			status: nil,
			want:   "",
		},
		{
			name:   "empty status map",
			status: &map[string]interface{}{},
			want:   "",
		},
		{
			name:   "phase key missing",
			status: &map[string]interface{}{"ready": true},
			want:   "",
		},
		{
			name:   "phase is nil",
			status: &map[string]interface{}{"phase": nil},
			want:   "",
		},
		{
			name:   "phase is non-string (number)",
			status: &map[string]interface{}{"phase": 42},
			want:   "",
		},
		{
			name:   "phase is non-string (map)",
			status: &map[string]interface{}{"phase": map[string]interface{}{"value": "RUNNING"}},
			want:   "",
		},
		{
			name:   "valid phase string",
			status: &map[string]interface{}{"phase": "RUNNING"},
			want:   "RUNNING",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := cmfsdk.ComputePool{Status: tt.status}
			if got := extractComputePoolPhase(pool); got != tt.want {
				t.Errorf("extractComputePoolPhase() = %q, want %q", got, tt.want)
			}
		})
	}
}
