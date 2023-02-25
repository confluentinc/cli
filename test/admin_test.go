package test

import (
	"os"
)

func (s *CLITestSuite) TestAdminPaymentDescribe() {
	tests := []CLITest{
		{args: "admin payment describe", fixture: "admin/payment/describe.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAdminPaymentDescribeMarketplaceOrg() {
	tests := []CLITest{
		{args: "admin payment describe", fixture: "admin/payment/describe-marketplace-org.golden"},
	}

	os.Setenv("IS_ORG_ON_MARKETPLACE", "true")
	defer unsetMarketplaceOrgEnv()

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAdminPaymentUpdate() {
	tests := []CLITest{
		{
			args:    "admin payment update",
			input:   "4242424242424242\n12/70\n999\nBrian Strauch\n",
			fixture: "admin/payment/update-success.golden",
		},
		{
			args:    "admin payment update", //testing with CVC failing regex check on first attempt
			input:   "4242424242424242\n12/70\n99\n999\nBrian Strauch\n",
			fixture: "admin/payment/update-bad-cvc.golden",
		},
	}
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAdminPromoAdd() {
	tests := []CLITest{
		{args: "admin promo add PROMOCODE", fixture: "admin/promo/add.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAdminPromoList() {
	tests := []CLITest{
		{args: "admin promo list", fixture: "admin/promo/list.golden"},
		{args: "admin promo list -o json", fixture: "admin/promo/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
