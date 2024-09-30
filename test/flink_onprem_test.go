package test

func (s *CLITestSuite) TestListFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/onprem/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment list-non-existent", fixture: "flink/onprem/application/list-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink application list --environment list-empty-environment", fixture: "flink/onprem/application/list-empty-env.golden"},
		{args: "flink application list --environment list-test  --output json", fixture: "flink/onprem/application/list-json.golden"},
		{args: "flink application list --environment list-test  --output human", fixture: "flink/onprem/application/list-human.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDeleteFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		// {args: "flink application delete test-app", fixture: "flink/onprem/application/delete-env-missing.golden", exitCode: 1},
		// {args: "flink application delete --environment delete-test", fixture: "flink/onprem/application/delete-missing-app.golden", exitCode: 1},
		{args: "flink application delete --force --environment delete-non-existent test-app", fixture: "flink/onprem/application/delete-non-existent-env.golden", exitCode: 1},
		{args: "flink application delete --environment delete-test delete-non-existent", fixture: "flink/onprem/application/delete-non-existent-app.golden", exitCode: 1},
		// success scenarios
		{args: "flink application delete --environment delete-test delete-test-app1 delete-test-app2", fixture: "flink/onprem/application/delete-success.golden"},
		// mixed scenarios
		{args: "flink application delete --environment delete-test delete-test-app1 delete-non-existent", fixture: "flink/onprem/application/delete-mixed.golden", exitCode: 1},
	}

	for _, test := range tests {
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
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDeleteFlinkEnvironments() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink environment delete", fixture: "flink/onprem/environment/delete-env-missing.golden", exitCode: 1},
		{args: "flink environment delete non-existent", fixture: "flink/onprem/environment/delete-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink environment delete delete-test delete-test2", fixture: "flink/onprem/environment/delete-success.golden"},
		// some failures and some successes
		{args: "flink environment delete delete-test delete-non-existent", fixture: "flink/onprem/environment/delete-mixed.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCreateFlinkApplication() {
	tests := []CLITest{
		{args: "flink application create --environment create-test test/fixtures/input/flink/onprem/application/create-new.json", fixture: "flink/onprem/application/create-success.golden"},
		{args: "flink application create --environment create-test test/fixtures/input/flink/onprem/application/create-unsuccessful-application.json", fixture: "flink/onprem/application/create-unsuccessful-application.golden", exitCode: 1},
		{args: "flink application create --environment create-test test/fixtures/input/flink/onprem/application/create-duplicate-application.json", fixture: "flink/onprem/application/create-duplicate-application.golden", exitCode: 1},
		{args: "flink application create --environment create-with-non-existent-environment test/fixtures/input/flink/onprem/application/create-with-non-existent-environment.json", fixture: "flink/onprem/application/create-with-non-existent-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUpdateFlinkApplication() {
	tests := []CLITest{
		{args: "flink application update --environment update-test test/fixtures/input/flink/onprem/application/update-successful.json", fixture: "flink/onprem/application/update-successful.golden"},
		{args: "flink application update --environment update-test test/fixtures/input/flink/onprem/application/update-non-existent.json", fixture: "flink/onprem/application/update-non-existent.golden", exitCode: 1},
		{args: "flink application update --environment update-test test/fixtures/input/flink/onprem/application/update-failure.json", fixture: "flink/onprem/application/update-failure.golden", exitCode: 1},
		{args: "flink application update --environment update-with-non-existent-environment test/fixtures/input/flink/onprem/application/update-with-non-existent-environment.json", fixture: "flink/onprem/application/update-with-non-existent-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.workflow = false
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCreateFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment create create-success", fixture: "flink/onprem/environment/create-success.golden"},
		{args: "flink environment create create-success-with-defaults --defaults test/fixtures/input/flink/onprem/environment/create-success-with-defaults.json", fixture: "flink/onprem/environment/create-success-with-defaults.golden"},
		// failure
		{args: "flink environment create create-failure", fixture: "flink/onprem/environment/create-failure.golden", exitCode: 1},
		{args: "flink environment create create-existing", fixture: "flink/onprem/environment/create-existing.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUpdateFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment update update-success --defaults '{\"property\": \"value\"}'", fixture: "flink/onprem/environment/update-success.golden"},
		// failure
		{args: "flink environment update update-failure", fixture: "flink/onprem/environment/update-failure.golden", exitCode: 1},
		{args: "flink environment update update-non-existent", fixture: "flink/onprem/environment/update-non-existent.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDescribeFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment describe describe-success", fixture: "flink/onprem/environment/describe-success.golden"},
		// failure
		{args: "flink environment describe describe-failure", fixture: "flink/onprem/environment/describe-failure.golden", exitCode: 1},
		{args: "flink environment describe describe-non-existent", fixture: "flink/onprem/environment/describe-non-existent.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDescribeFlinkApplication() {
	tests := []CLITest{
		// success
		{args: "flink application describe --environment describe-test describe-success", fixture: "flink/onprem/application/describe-success.golden"},
		// failure
		{args: "flink application describe  --environment describe-test describe-failure", fixture: "flink/onprem/application/describe-failure.golden", exitCode: 1},
		{args: "flink application describe --environment describe-test describe-non-existent", fixture: "flink/onprem/application/describe-non-existent.golden", exitCode: 1},
		{args: "flink application describe --environment describe-non-existent describe-non-existent-environment", fixture: "flink/onprem/application/describe-non-existent-environment.golden", exitCode: 1},
		{args: "flink application describe describe-no-environment", fixture: "flink/onprem/environment/describe-no-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}
