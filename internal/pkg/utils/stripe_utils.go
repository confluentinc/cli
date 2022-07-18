package utils

import (
	"errors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/token"
)

const (
	stripeTestKey = "pk_test_0MJU6ihIFpxuWMwG6HhjGQ8P"
	stripeLiveKey = "pk_live_t0P8AKi9DEuvAqfKotiX5xHM"
)

func setStripeKey(isTest bool) {
	if isTest {
		stripe.Key = stripeTestKey
	} else {
		stripe.Key = stripeLiveKey
	}
	stripe.DefaultLeveledLogger = &stripe.LeveledLogger{
		Level: 0,
	}
}

// hasDefaultPaymentMethod reuses the logic of HasDefaultPaymentMethod defined in https://github.com/confluentinc/cc-billing-worker/blob/master/stripe/customer.go
func hasDefaultPaymentMethod(stripeCustomerId string, isTest bool) (bool, error) {
	setStripeKey(isTest)

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

func NewStripeToken(cardNumber string, expMonth string, expYear string, cvc string, name string, isTest bool) (*stripe.Token, error) {
	setStripeKey(isTest)

	params := &stripe.TokenParams{
		Card: &stripe.CardParams{
			Number:   stripe.String(cardNumber),
			ExpMonth: stripe.String(expMonth),
			ExpYear:  stripe.String(expYear),
			CVC:      stripe.String(cvc),
			Name:     stripe.String(name),
		},
	}

	stripeToken, err := token.New(params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			return nil, errors.New(stripeErr.Msg)
		}
		return nil, err
	}

	return stripeToken, nil
}
