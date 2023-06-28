package test

func (s *CLITestSuite) TestAdminPaymentDescribe() {
	tests := []CLITest{
		{args: "admin payment describe", fixture: "admin/payment/describe.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAdminPaymentDescribe_MarketplaceOrg() {
	tests := []CLITest{
		{args: "admin payment describe", fixture: "admin/payment/describe-marketplace-org.golden"},
	}

	s.T().Setenv("IS_ORG_ON_MARKETPLACE", "true")

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
			fixture: "admin/payment/update.golden",
		},
		{
			args:    "admin payment update",
			input:   "bad card number\n4242424242424242\nbad expiration\n12/70\nbad cvc\n999\nBrian Strauch\n",
			fixture: "admin/payment/update-retry.golden",
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
