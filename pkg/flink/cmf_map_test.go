package flink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMapField(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name   string
			m      map[string]any
			want   string
			wantOk bool
		}{
			{name: "nil map", m: nil, want: "", wantOk: false},
			{name: "empty map", m: map[string]any{}, want: "", wantOk: false},
			{name: "key missing", m: map[string]any{"other": "x"}, want: "", wantOk: false},
			{name: "value nil", m: map[string]any{"phase": nil}, want: "", wantOk: false},
			{name: "string value", m: map[string]any{"phase": "RUNNING"}, want: "RUNNING", wantOk: true},
			{name: "empty string", m: map[string]any{"phase": ""}, want: "", wantOk: true},
			{name: "wrong type int", m: map[string]any{"phase": 42}, want: "", wantOk: false},
			{name: "wrong type bool", m: map[string]any{"phase": true}, want: "", wantOk: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, ok := GetMapField[string](tt.m, "phase", "test")
				require.Equal(t, tt.wantOk, ok)
				require.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("int64", func(t *testing.T) {
		got, ok := GetMapField[int64](map[string]any{"observedGeneration": int64(7)}, "observedGeneration", "test")
		require.True(t, ok)
		require.Equal(t, int64(7), got)

		// JSON decodes numbers as float64, not int64.
		_, ok = GetMapField[int64](map[string]any{"observedGeneration": float64(7)}, "observedGeneration", "test")
		require.False(t, ok)
	})

	t.Run("nested map", func(t *testing.T) {
		nested := map[string]any{"jobName": "foo", "state": "RUNNING"}
		root := map[string]any{"jobStatus": nested}

		got, ok := GetMapField[map[string]any](root, "jobStatus", "test")
		require.True(t, ok)
		require.Equal(t, nested, got)

		state, ok := GetMapField[string](got, "state", "test")
		require.True(t, ok)
		require.Equal(t, "RUNNING", state)
	})

	t.Run("slice", func(t *testing.T) {
		conditions := []any{"Ready", "Healthy"}
		got, ok := GetMapField[[]any](map[string]any{"conditions": conditions}, "conditions", "test")
		require.True(t, ok)
		require.Equal(t, conditions, got)
	})
}
