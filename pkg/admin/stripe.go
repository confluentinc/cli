package admin

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/token"
)

const (
	stripeLiveKey = "pk_live_t0P8AKi9DEuvAqfKotiX5xHM"
	stripeTestKey = "pk_test_0MJU6ihIFpxuWMwG6HhjGQ8P"
)

func NewStripeToken(cfg *config.Config, number, expiration, cvc, name string) (*stripe.Token, error) {
	stripe.DefaultLeveledLogger = &stripe.LeveledLogger{Level: stripe.LevelNull}

	if cfg.Context().GetPlatform().GetName() == "confluent.cloud" {
		stripe.Key = stripeLiveKey
	} else {
		stripe.Key = stripeTestKey
	}

	exp := strings.Split(expiration, "/")

	params := &stripe.TokenParams{Card: &stripe.CardParams{
		Number:   stripe.String(number),
		ExpMonth: stripe.String(exp[0]),
		ExpYear:  stripe.String(exp[1]),
		CVC:      stripe.String(cvc),
		Name:     stripe.String(name),
	}}

	stripeToken, err := token.New(params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			return nil, fmt.Errorf(stripeErr.Msg)
		}
		return nil, err
	}

	return stripeToken, nil
}
