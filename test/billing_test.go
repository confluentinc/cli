package test

func (s *CLITestSuite) TestCostList() {
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
