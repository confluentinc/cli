package utils

import (
	"os"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

const (
	stripeTestKey = "pk_test_0MJU6ihIFpxuWMwG6HhjGQ8P"
	stripeLiveKey = "pk_live_t0P8AKi9DEuvAqfKotiX5xHM"
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
		hasPaymentMethod := os.Getenv("HAS_PAYMENT_METHOD")
		if hasPaymentMethod == "true" {
			return false
		} else {
			return true
		}
	}

	if orgv1.BillingMethod_STRIPE == org.GetPlan().GetBilling().GetMethod() {
		stripeCustomerId := org.GetPlan().GetBilling().GetStripeCustomerId()
		// If there is no stripe customer ID, then the org does not have payment info
		if "" == stripeCustomerId {
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

// hasDefaultPaymentMethod reuses the logic of HasDefaultPaymentMethod defined in https://github.com/confluentinc/cc-billing-worker/blob/master/stripe/customer.go
func hasDefaultPaymentMethod(stripeCustomerId string, isTest bool) (bool, error) {
	if isTest {
		stripe.Key = stripeTestKey
	} else {
		stripe.Key = stripeLiveKey
	}
	stripe.DefaultLeveledLogger = &stripe.LeveledLogger{
		Level: 0,
	}

	stripeCustomer, err := customer.Get(stripeCustomerId, nil)
	if err != nil {
		return false, err
	}

	if stripeCustomer == nil {
		return false, errors.New("customer is nil")
	}

	if stripeCustomer.DefaultSource != nil {
		return true, nil
	}

	if stripeCustomer.InvoiceSettings != nil && stripeCustomer.InvoiceSettings.DefaultPaymentMethod != nil {
		return true, nil
	}

	return false, nil
}
