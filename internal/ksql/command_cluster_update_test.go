package ksql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCsuForUpdate(t *testing.T) {
	tests := []struct {
		name        string
		csu         int32
		expectErr   bool
		errContains string
	}{
		{name: "valid 4", csu: 4},
		{name: "valid 8", csu: 8},
		{name: "valid 12", csu: 12},
		{name: "valid 16", csu: 16},
		{name: "valid 20", csu: 20},
		{name: "valid 24", csu: 24},
		{name: "valid 28", csu: 28},
		{
			name:        "legacy size 1 rejected",
			csu:         1,
			expectErr:   true,
			errContains: "not a valid CSU size",
		},
		{
			name:        "legacy size 2 rejected",
			csu:         2,
			expectErr:   true,
			errContains: "not a valid CSU size",
		},
		{
			name:        "in-range but non-canonical (5) rejected",
			csu:         5,
			expectErr:   true,
			errContains: "not a valid CSU size",
		},
		{
			name:        "in-range but non-canonical (10) rejected",
			csu:         10,
			expectErr:   true,
			errContains: "not a valid CSU size",
		},
		{
			name:        "above 28 routes to support-ticket message",
			csu:         32,
			expectErr:   true,
			errContains: "support ticket",
		},
		{
			name:        "well above ceiling routes to support-ticket message",
			csu:         128,
			expectErr:   true,
			errContains: "support ticket",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCsuForUpdate(tc.csu)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFormatCsuList(t *testing.T) {
	require.Equal(t, "4, 8, 12, 16, 20, 24, 28", formatCsuList(validCsuSizes))
	// Input order should not matter; output is sorted ascending.
	require.Equal(t, "4, 8, 16", formatCsuList([]int32{16, 4, 8}))
}
