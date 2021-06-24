package test

func (s *CLITestSuite) TestPromoAdd() {
	tests := []CLITest{
		{
			Args:    "admin promo add PROMOCODE",
			Fixture: "admin/promo-add.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}

func (s *CLITestSuite) TestPromoList() {
	tests := []CLITest{
		{
			Args:    "admin promo list",
			Fixture: "admin/promo-list.golden",
		},
		{
			Args:    "admin promo list -o json",
			Fixture: "admin/promo-list-json.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}
