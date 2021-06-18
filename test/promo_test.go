package test

func (s *CLITestSuite) TestPromoAdd() {
	tests := []CLITest{
		{
			args:    "admin promo add PROMOCODE",
			fixture: "admin/promo-add.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestPromoList() {
	tests := []CLITest{
		{
			args:    "admin promo list",
			fixture: "admin/promo-list.golden",
		},
		{
			args:    "admin promo list -o json",
			fixture: "admin/promo-list-json.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
