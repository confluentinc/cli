package utils

import (
	"os"

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

// IsOrgOnFreeTrial reuses the logic of IsOnFreeTrial defined in https://github.com/confluentinc/cc-billing-worker/blob/master/handler/org_cluster_handler.go
func IsOrgOnFreeTrial(org *orgv1.Organization, isTest bool) bool {
	// If the organization is deactivated, then it is not on the free trial
	if org.GetDeactivated() {
		return false
	}

	// If the organization is currently suspended because of the end of the free trial, then they are on the free trial
	if orgv1.SuspensionStatusType_SUSPENSION_IN_PROGRESS == org.GetSuspensionStatus().GetStatus() ||
		orgv1.SuspensionStatusType_SUSPENSION_COMPLETED == org.GetSuspensionStatus().GetStatus() {
		return orgv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL == org.GetSuspensionStatus().GetEventType()
	}

	if isTest {
		return os.Getenv("HAS_PAYMENT_METHOD") != "true"
	}

	if orgv1.BillingMethod_STRIPE == org.GetPlan().GetBilling().GetMethod() {
		stripeCustomerId := org.GetPlan().GetBilling().GetStripeCustomerId()
		// If there is no stripe customer ID, then the org does not have payment info
		if stripeCustomerId == "" {
			return true
		}

		hasDefaultPaymentMethod, err := hasDefaultPaymentMethod(stripeCustomerId, isTest)
		if err != nil {
			return false
		}

		// If Stripe does not have an associated payment method
		return !hasDefaultPaymentMethod
	}

	return false
}
