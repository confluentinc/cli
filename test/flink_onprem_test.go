package test

func (s *CLITestSuite) TestListFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/onprem/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment non-existent", fixture: "flink/onprem/application/list-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink application list --environment empty-environment", fixture: "flink/onprem/application/list-empty-env.golden"},
		{args: "flink application list --environment test  --output json", fixture: "flink/onprem/application/list-json.golden"},
		{args: "flink application list --environment test  --output human", fixture: "flink/onprem/application/list-human.golden"},
	}

	for _, test := range tests {
		test.workflow = false
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDeleteFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application delete test-app", fixture: "flink/onprem/application/delete-env-missing.golden", exitCode: 1},
		{args: "flink application delete --environment defauklt", fixture: "flink/onprem/application/delete-missing-app.golden", exitCode: 1},
		{args: "flink application delete --environment non-existent test-app", fixture: "flink/onprem/application/delete-non-existent-env.golden"},
		{args: "flink application delete --environment default non-existent", fixture: "flink/onprem/application/delete-non-existent-app.golden"},
		// success scenarios
		{args: "flink application delete --environment default test,test-app", fixture: "flink/onprem/application/delete-success.golden"},
		// mixed scnearios
		{args: "flink application delete --environment default test,non-existent", fixture: "flink/onprem/application/delete-mixed.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
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

func (s *CLITestSuite) TestDeleteFlinkEnvironments() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink environment delete", fixture: "flink/onprem/environment/delete-env-missing.golden", exitCode: 1},
		{args: "flink environment delete non-existent", fixture: "flink/onprem/environment/delete-non-existent-env.golden"},
		// success scenarios
		{args: "flink environment delete test,test2", fixture: "flink/onprem/environment/delete-success.golden"},
		// some failures and some successes
		{args: "flink environment delete test,non-existent", fixture: "flink/onprem/environment/delete-mixed.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		test.workflow = false
		s.runIntegrationTest(test)
	}
}
