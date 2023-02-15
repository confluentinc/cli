package utils

import (
	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
)

func IsOrgSuspended(suspensionStatus *ccloudv1.SuspensionStatus) bool {
	status := suspensionStatus.GetStatus()
	return status == ccloudv1.SuspensionStatusType_SUSPENSION_IN_PROGRESS || status == ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED
}

func IsOrgEndOfFreeTrialSuspended(suspensionStatus *ccloudv1.SuspensionStatus) bool {
	eventType := suspensionStatus.GetEventType()
	return IsOrgSuspended(suspensionStatus) && eventType == ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL
}

func IsLoginBlockedByOrgSuspension(suspensionStatus *ccloudv1.SuspensionStatus) bool {
	eventType := suspensionStatus.GetEventType()
	return IsOrgSuspended(suspensionStatus) && eventType != ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL
}
