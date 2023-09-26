package utils

import (
	"errors"
	"strings"

	"github.com/stripe/stripe-go"
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
	stripe.DefaultLeveledLogger = &stripe.LeveledLogger{Level: 0}
}

func NewStripeToken(cardNumber, expiration, cvc, name string, isTest bool) (*stripe.Token, error) {
	setStripeKey(isTest)

	exp := strings.Split(expiration, "/")

	params := &stripe.TokenParams{
		Card: &stripe.CardParams{
			Number:   stripe.String(cardNumber),
			ExpMonth: stripe.String(exp[0]),
			ExpYear:  stripe.String(exp[1]),
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
