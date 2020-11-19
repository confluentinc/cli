package test

import (
	"encoding/json"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func (s *CLITestSuite) TestPaymentDescribe() {
	tests := []CLITest{
		{
			args:    "admin payment describe",
			fixture: "admin/payment-describe.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func (s *CLITestSuite) TestPaymentUpdate() {
	tests := []CLITest{
		{
			args:    	"admin payment update",
			cmdFuncs: 	[]cmdFunc{stdinPipeFunc(strings.NewReader("4242424242424242\n12/70\n999\nBrian Strauch\n"))},
			fixture: 	"admin/payment-update-success.golden",
		},
	}
	for _, test := range tests {
		test.login = "default"
		loginUrl := serve(s.T(), "").URL
		s.runCcloudTest(test, loginUrl)
	}
}

func handlePaymentInfo(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := orgv1.GetPaymentInfoReply{
			Card:                 &orgv1.Card{
				Cardholder:           "Miles Todzo",
				Brand:                "Visa",
				Last4:                "4242",
				ExpMonth:             "01",
				ExpYear:              "99",
			},
			Organization:         &orgv1.Organization{
				Id:                      0,
			},
			Error:                nil,
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}
