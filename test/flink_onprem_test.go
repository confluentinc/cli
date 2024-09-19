package test

func (s *CLITestSuite) TestListFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/onprem/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment non-existent", fixture: "flink/onprem/application/list-non-existent-env.golden", exitCode: 1},
		{args: "flink application list --environment empty-environment", fixture: "flink/onprem/application/list-empty-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink application list --environment test  --output json", fixture: "flink/onprem/application/list-json.golden"},
		{args: "flink application list --environment test  --output human", fixture: "flink/onprem/application/list-human.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		test.workflow = false
		s.runIntegrationTest(test)
	}
}
