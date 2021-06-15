package test

func (s *CLITestSuite) TestPromoAdd() {
	tests := []CLITest{
		{
			args:    "admin promo add XXXXX",
			fixture: "admin/promo-add.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
