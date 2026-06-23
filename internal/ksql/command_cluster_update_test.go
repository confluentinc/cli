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

// TestBuildUpdateLongDescription pins the customer-facing help text. KSQL-15168
// rewrote this to advertise both expansion and shrink (and the TERMINATE
// remediation when the server refuses a shrink). A future change that
// reverts to expand-only wording, drops the TERMINATE guidance, or drops
// the valid CSU listing would break this test.
func TestBuildUpdateLongDescription(t *testing.T) {
	long := buildUpdateLongDescription()

	require.Contains(t, long, "Both expansion (increase) and shrink (decrease) are supported",
		"long description must advertise both directions")
	require.Contains(t, long, "TERMINATE <query_id>",
		"long description must surface the customer-side remediation for a refused shrink")
	require.Contains(t, long, "4, 8, 12, 16, 20, 24, 28",
		"long description must enumerate valid CSU sizes (kept in sync with validCsuSizes)")
	require.Contains(t, long, "rolling restart",
		"long description must call out the rolling-restart behavior")
}

// TestCsuSupportTicketMessage pins the support-ticket fallback message.
// Customer-visible string; a regression in wording would silently change
// the experience for anyone passing a CSU above maxSelfServeCSU.
func TestCsuSupportTicketMessage(t *testing.T) {
	msg := csuSupportTicketMessage()
	require.Contains(t, msg, "support ticket")
	require.Contains(t, msg, "28", "message must name the self-serve ceiling")
}
