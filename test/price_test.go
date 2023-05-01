package test

func (s *CLITestSuite) TestPriceList() {
	tests := []CLITest{
		{args: "price list --cloud aws --region us-east-1", fixture: "price/list.golden"},
		{args: "price list --cloud aws --region us-east-1 -o json", fixture: "price/list-json.golden"},
		{args: "price list --cloud aws --region us-east-1 --legacy", fixture: "price/list-legacy.golden"},
		{args: `price list --cloud aws --region ""`, fixture: "price/list-empty-flag.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
