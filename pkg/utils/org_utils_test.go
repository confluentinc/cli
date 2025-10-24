package utils

import (
	"testing"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/stretchr/testify/require"
)

func TestIsOrgPauseTrialSuspended(t *testing.T) {
	tests := []struct {
		name             string
		suspensionStatus *ccloudv1.SuspensionStatus
		expected         bool
	}{
		{
			name: "PAUSE_TRIAL event should be pause trial",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_PAUSE_TRIAL,
			},
			expected: true,
		},
		{
			name: "END_OF_FREE_TRIAL event should not be pause trial",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL,
			},
			expected: false,
		},
		{
			name: "CUSTOMER_INITIATED_ORG_DEACTIVATION should not be pause trial",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION,
			},
			expected: false,
		},
		{
			name: "Not suspended should not be pause trial",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_UNKNOWN,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_PAUSE_TRIAL,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOrgPauseTrialSuspended(tt.suspensionStatus)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLoginBlockedByOrgSuspension(t *testing.T) {
	tests := []struct {
		name             string
		suspensionStatus *ccloudv1.SuspensionStatus
		expected         bool
	}{
		{
			name: "PAUSE_TRIAL should not block login",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_PAUSE_TRIAL,
			},
			expected: false,
		},
		{
			name: "END_OF_FREE_TRIAL should not block login",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL,
			},
			expected: false,
		},
		{
			name: "CUSTOMER_INITIATED_ORG_DEACTIVATION should block login",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION,
			},
			expected: true,
		},
		{
			name: "Not suspended should not block login",
			suspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_UNKNOWN,
				EventType: ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLoginBlockedByOrgSuspension(tt.suspensionStatus)
			require.Equal(t, tt.expected, result)
		})
	}
}
