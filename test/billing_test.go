package test

func (s *CLITestSuite) TestCostList() {
	tests := []CLITest{
		{args: "billing cost list --start-date 2023-01-01 --end-date 2023-01-02", fixture: "billing/list.golden"},
		{args: "billing cost list --start-date 2023-01-01 --end-date 2023-01-02 -o json", fixture: "billing/list_json.golden"},
		{args: "billing cost list --start-date 2023-01-01 --end-date 2023-01-02 -o yaml", fixture: "billing/list_yaml.golden"},
		{args: "billing cost list --start-date 2023-01-01", fixture: "billing/list_missing_arg.golden", exitCode: 1},
		{args: "billing cost list --start-date 01-02-2023 --end-date 01-01-2023", fixture: "billing/list_invalid_format.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
