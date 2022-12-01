package utils

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

func IsOrgSuspended(suspensionStatus *orgv1.SuspensionStatus) bool {
	status := suspensionStatus.GetStatus()
	return status == orgv1.SuspensionStatusType_SUSPENSION_IN_PROGRESS || status == orgv1.SuspensionStatusType_SUSPENSION_COMPLETED
}

func IsOrgEndOfFreeTrialSuspended(suspensionStatus *orgv1.SuspensionStatus) bool {
	eventType := suspensionStatus.GetEventType()
	return IsOrgSuspended(suspensionStatus) && eventType == orgv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL
}

func IsLoginBlockedByOrgSuspension(suspensionStatus *orgv1.SuspensionStatus) bool {
	eventType := suspensionStatus.GetEventType()
	return IsOrgSuspended(suspensionStatus) && eventType != orgv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL
}
