package test

func (s *CLITestSuite) TestListFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/onprem/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment non-existent", fixture: "flink/onprem/application/list-non-existent-env.golden", exitCode: 1},
		// // success scenarios
		{args: "flink application list --environment empty-environment", fixture: "flink/onprem/application/list-empty-env.golden"},
		{args: "flink application list --environment test  --output json", fixture: "flink/onprem/application/list-json.golden"},
		{args: "flink application list --environment test  --output human", fixture: "flink/onprem/application/list-human.golden"},
	}

	for _, test := range tests {
		test.workflow = false
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestListFlinkEnvironments() {
	tests := []CLITest{
		// success scenarios
		{args: "flink environment list --output json", fixture: "flink/onprem/environment/list-json.golden"},
		{args: "flink environment list --output human", fixture: "flink/onprem/environment/list-human.golden"},
	}

	for _, test := range tests {
		test.workflow = false
		s.runIntegrationTest(test)
	}
}
