package test

func (s *CLITestSuite) TestCostList() {
	tests := []CLITest{
		{args: "billing cost list 2023-01-01 2023-01-02", fixture: "billing/list.golden"},
		{args: "billing cost list 2023-01-01 2023-01-02 -o json", fixture: "billing/list_json.golden"},
		{args: "billing cost list 2023-01-01 2023-01-02 -o yaml", fixture: "billing/list_yaml.golden"},
		{args: "billing cost list 2023-01-01", fixture: "billing/list_missing_arg.golden", exitCode: 1},
		{args: "billing cost list 2023-01-01", fixture: "billing/list_invalid_format.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
