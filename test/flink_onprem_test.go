package test

import "os"

func (s *CLITestSuite) TestListFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment non-existent", fixture: "flink/application/list-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink application list --environment test", fixture: "flink/application/list-empty-env.golden"},
		{args: "flink application list --environment default  --output json", fixture: "flink/application/list-json.golden"},
		{args: "flink application list --environment default  --output human", fixture: "flink/application/list-human.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDeleteFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application delete default-application-1", fixture: "flink/application/delete-env-missing.golden", exitCode: 1},
		{args: "flink application delete --environment default", fixture: "flink/application/delete-missing-app.golden", exitCode: 1},
		{args: "flink application delete --force --environment non-existent default-application-1", fixture: "flink/application/delete-non-existent-env.golden", exitCode: 1},
		{args: "flink application delete --environment default non-existent", fixture: "flink/application/delete-non-existent-app.golden", exitCode: 1},
		// success scenarios
		{args: "flink application delete --environment default default-application-1 default-application-2 --force", fixture: "flink/application/delete-success.golden"},
		// mixed scenarios
		{args: "flink application delete --environment delete-test default-application-1 non-existent --force", fixture: "flink/application/delete-mixed.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestListFlinkEnvironments() {
	tests := []CLITest{
		// success scenarios
		{args: "flink environment list --output json", fixture: "flink/environment/list-json.golden"},
		{args: "flink environment list --output human", fixture: "flink/environment/list-human.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDeleteFlinkEnvironments() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink environment delete", fixture: "flink/environment/delete-env-missing.golden", exitCode: 1},
		{args: "flink environment delete non-existent", fixture: "flink/environment/delete-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink environment delete default test --force", fixture: "flink/environment/delete-success.golden"},
		// some failures and some successes
		{args: "flink environment delete default non-existent --force", fixture: "flink/environment/delete-mixed.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCreateFlinkApplication() {
	tests := []CLITest{
		// failure
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-unsuccessful-application.json", fixture: "flink/application/create-unsuccessful-application.golden", exitCode: 1},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-duplicate-application.json", fixture: "flink/application/create-duplicate-application.golden", exitCode: 1},
		{args: "flink application create --environment non-existent test/fixtures/input/flink/application/create-with-non-existent-environment.json", fixture: "flink/application/create-with-non-existent-environment.golden", exitCode: 1},
		// success
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json", fixture: "flink/application/create-success.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUpdateFlinkApplication() {
	tests := []CLITest{
		// failure
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-non-existent.json", fixture: "flink/application/update-non-existent.golden", exitCode: 1},
		{args: "flink application update --environment update-failure test/fixtures/input/flink/application/update-failure.json", fixture: "flink/application/update-failure.golden", exitCode: 1},
		{args: "flink application update --environment non-existent test/fixtures/input/flink/application/update-with-non-existent-environment.json", fixture: "flink/application/update-with-non-existent-environment.golden", exitCode: 1},
		// success
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json", fixture: "flink/application/update-successful.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCreateFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment create default-2 --kubernetes-namespace default-staging", fixture: "flink/environment/create-success.golden"},
		{args: "flink environment create default-2 --defaults test/fixtures/input/flink/environment/create-success-with-defaults.json --kubernetes-namespace default-staging", fixture: "flink/environment/create-success-with-defaults.golden"},
		// failure
		{args: "flink environment create default-failure --kubernetes-namespace default-staging", fixture: "flink/environment/create-failure.golden", exitCode: 1},
		{args: "flink environment create default --kubernetes-namespace default-staging", fixture: "flink/environment/create-existing.golden", exitCode: 1},
		{args: "flink environment create default", fixture: "flink/environment/create-no-namespace.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUpdateFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment update default --defaults '{\"property\": \"value\"}'", fixture: "flink/environment/update-success.golden"},
		// failure
		{args: "flink environment update update-failure", fixture: "flink/environment/update-failure.golden", exitCode: 1},
		{args: "flink environment update non-existent", fixture: "flink/environment/update-non-existent.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDescribeFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment describe default", fixture: "flink/environment/describe-success.golden"},
		{args: "flink environment describe default --output json", fixture: "flink/environment/describe-success-json.golden"},
		{args: "flink environment describe default --output yaml", fixture: "flink/environment/describe-success-yaml.golden"},
		// failure
		{args: "flink environment describe non-existent", fixture: "flink/environment/describe-non-existent.golden", exitCode: 1},
		{args: "flink environment describe", fixture: "flink/environment/describe-no-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDescribeFlinkApplication() {
	tests := []CLITest{
		// success
		{args: "flink application describe --environment default default-application-1", fixture: "flink/application/describe-success.golden"},
		{args: "flink application describe --environment default default-application-1 --output yaml", fixture: "flink/application/describe-success-yaml.golden"},
		// failure
		{args: "flink application describe --environment default non-existent", fixture: "flink/application/describe-non-existent.golden", exitCode: 1},
		{args: "flink application describe --environment non-existent default-application", fixture: "flink/application/describe-non-existent-environment.golden", exitCode: 1},
		{args: "flink application describe default-application-1", fixture: "flink/application/describe-no-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestForwardWebUiFlinkApplication() {
	// We cannot test the success cases as they require a running CMF service. However we can test some basic failure cases.
	tests := []CLITest{
		// failure
		{args: "flink --url dummy-url application web-ui-forward forward-negative-port --environment forward-test --port -30", fixture: "flink/application/forward-negative-port.golden", exitCode: 1},
		{args: "flink --url dummy-url application web-ui-forward non-existent --environment default", fixture: "flink/application/forward-nonexistent-application.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUseEnvironmentFlinkApplications() {
	// The test harness already creates a "temporary" home directory for us to use, but it is the same across tests.
	// So, we need to change $HOME.

	tmpDir, err := os.MkdirTemp("", "environment-use-test-home")
	if err != nil {
		s.T().Fatalf("Failed to create temporary directory: %v", err)
	}
	// we don't need to change env with a $HOME variable.
	tests := []CLITest{
		// failure
		{args: "flink environment use non-existent", fixture: "flink/environment/use-non-existent.golden", exitCode: 1},
		// success, followed by listing applications
		{args: "flink environment use default", fixture: "flink/environment/use-success.golden", env: []string{"HOME=" + tmpDir}},
		{args: "flink application describe default-application-1", fixture: "flink/application/describe-after-use-success.golden", env: []string{"HOME=" + tmpDir}},
	}

	for _, test := range tests {
		// Don't reset state between tests
		test.workflow = true
		s.runIntegrationTest(test)
	}
}
