package test

func (s *CLITestSuite) TestBillingCostList() {
	tests := []CLITest{
		{args: "billing cost list --start-date 2023-01-01 --end-date 2023-01-02", fixture: "billing/cost/list.golden"},
		{args: "billing cost list --start-date 2023-01-01 --end-date 2023-01-02 -o json", fixture: "billing/cost/list-json.golden"},
		{args: "billing cost list --start-date 2023-01-01 --end-date 2023-01-02 -o yaml", fixture: "billing/cost/list-yaml.golden"},
		{args: "billing cost list --start-date 2023-01-01", fixture: "billing/cost/list-missing-flag.golden", exitCode: 1},
		{args: "billing cost list --start-date 01-02-2023 --end-date 01-01-2023", fixture: "billing/cost/list-invalid-format.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestBillingPriceList() {
	tests := []CLITest{
		{args: "billing price list --cloud aws --region us-east-1", fixture: "billing/price/list.golden"},
		{args: "billing price list --cloud aws --region us-east-1 -o json", fixture: "billing/price/list-json.golden"},
		{args: "billing price list --cloud aws --region us-east-1 --legacy", fixture: "billing/price/list-legacy.golden"},
		{args: `billing price list --cloud aws --region ""`, fixture: "billing/price/list-empty-flag.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestBillingPaymentDescribe() {
	tests := []CLITest{
		{args: "billing payment describe", fixture: "billing/payment/describe.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestBillingPaymentDescribe_MarketplaceOrg() {
	tests := []CLITest{
		{args: "billing payment describe", fixture: "billing/payment/describe-marketplace-org.golden"},
	}

	s.T().Setenv("IS_ORG_ON_MARKETPLACE", "true")

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestBillingPaymentUpdate() {
	tests := []CLITest{
		{
			args:    "billing payment update",
			input:   "4242424242424242\n12/70\n999\nBrian Strauch\n",
			fixture: "billing/payment/update.golden",
		},
		{
			args:    "billing payment update",
			input:   "bad card number\n4242424242424242\nbad expiration\n12/70\nbad cvc\n999\nBrian Strauch\n",
			fixture: "billing/payment/update-retry.golden",
		},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestBillingPromoAdd() {
	tests := []CLITest{
		{args: "billing promo add PROMOCODE", fixture: "billing/promo/add.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestBillingPromoList() {
	tests := []CLITest{
		{args: "billing promo list", fixture: "billing/promo/list.golden"},
		{args: "billing promo list -o json", fixture: "billing/promo/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
