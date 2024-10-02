package test

import "os"

func (s *CLITestSuite) TestListFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/onprem/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment non-existent", fixture: "flink/onprem/application/list-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink application list --environment test", fixture: "flink/onprem/application/list-empty-env.golden"},
		{args: "flink application list --environment default  --output json", fixture: "flink/onprem/application/list-json.golden"},
		{args: "flink application list --environment default  --output human", fixture: "flink/onprem/application/list-human.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDeleteFlinkApplications() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application delete default-application-1", fixture: "flink/onprem/application/delete-env-missing.golden", exitCode: 1},
		{args: "flink application delete --environment default", fixture: "flink/onprem/application/delete-missing-app.golden", exitCode: 1},
		{args: "flink application delete --force --environment non-existent default-application-1", fixture: "flink/onprem/application/delete-non-existent-env.golden", exitCode: 1},
		{args: "flink application delete --environment default non-existent", fixture: "flink/onprem/application/delete-non-existent-app.golden", exitCode: 1},
		// success scenarios
		{args: "flink application delete --environment default default-application-1 default-application-2 --force", fixture: "flink/onprem/application/delete-success.golden"},
		// mixed scenarios
		{args: "flink application delete --environment delete-test default-application-1 non-existent --force", fixture: "flink/onprem/application/delete-mixed.golden", exitCode: 1},
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
		{args: "flink environment delete default test --force", fixture: "flink/onprem/environment/delete-success.golden"},
		// some failures and some successes
		{args: "flink environment delete default non-existent --force", fixture: "flink/onprem/environment/delete-mixed.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCreateFlinkApplication() {
	tests := []CLITest{
		// failure
		{args: "flink application create --environment default test/fixtures/input/flink/onprem/application/create-unsuccessful-application.json", fixture: "flink/onprem/application/create-unsuccessful-application.golden", exitCode: 1},
		{args: "flink application create --environment default test/fixtures/input/flink/onprem/application/create-duplicate-application.json", fixture: "flink/onprem/application/create-duplicate-application.golden", exitCode: 1},
		{args: "flink application create --environment non-existent test/fixtures/input/flink/onprem/application/create-with-non-existent-environment.json", fixture: "flink/onprem/application/create-with-non-existent-environment.golden", exitCode: 1},
		// success
		{args: "flink application create --environment default test/fixtures/input/flink/onprem/application/create-new.json", fixture: "flink/onprem/application/create-success.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUpdateFlinkApplication() {
	tests := []CLITest{
		// failure
		{args: "flink application update --environment default test/fixtures/input/flink/onprem/application/update-non-existent.json", fixture: "flink/onprem/application/update-non-existent.golden", exitCode: 1},
		{args: "flink application update --environment update-failure test/fixtures/input/flink/onprem/application/update-failure.json", fixture: "flink/onprem/application/update-failure.golden", exitCode: 1},
		{args: "flink application update --environment non-existent test/fixtures/input/flink/onprem/application/update-with-non-existent-environment.json", fixture: "flink/onprem/application/update-with-non-existent-environment.golden", exitCode: 1},
		// success
		{args: "flink application update --environment default test/fixtures/input/flink/onprem/application/update-successful.json", fixture: "flink/onprem/application/update-successful.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCreateFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment create default-2", fixture: "flink/onprem/environment/create-success.golden"},
		{args: "flink environment create default-2 --defaults test/fixtures/input/flink/onprem/environment/create-success-with-defaults.json", fixture: "flink/onprem/environment/create-success-with-defaults.golden"},
		// failure
		{args: "flink environment create default-failure", fixture: "flink/onprem/environment/create-failure.golden", exitCode: 1},
		{args: "flink environment create default", fixture: "flink/onprem/environment/create-existing.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUpdateFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment update default --defaults '{\"property\": \"value\"}'", fixture: "flink/onprem/environment/update-success.golden"},
		// failure
		{args: "flink environment update update-failure", fixture: "flink/onprem/environment/update-failure.golden", exitCode: 1},
		{args: "flink environment update non-existent", fixture: "flink/onprem/environment/update-non-existent.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDescribeFlinkEnvironment() {
	tests := []CLITest{
		// success
		{args: "flink environment describe default", fixture: "flink/onprem/environment/describe-success.golden"},
		{args: "flink environment describe default --output json", fixture: "flink/onprem/environment/describe-success-json.golden"},
		// failure
		{args: "flink environment describe non-existent", fixture: "flink/onprem/environment/describe-non-existent.golden", exitCode: 1},
		{args: "flink environment describe", fixture: "flink/onprem/environment/describe-no-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestDescribeFlinkApplication() {
	tests := []CLITest{
		// success
		{args: "flink application describe --environment default default-application-1", fixture: "flink/onprem/application/describe-success.golden"},
		// failure
		{args: "flink application describe --environment default non-existent", fixture: "flink/onprem/application/describe-non-existent.golden", exitCode: 1},
		{args: "flink application describe --environment non-existent default-application", fixture: "flink/onprem/application/describe-non-existent-environment.golden", exitCode: 1},
		{args: "flink application describe default-application-1", fixture: "flink/onprem/application/describe-no-environment.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestForwardWebUiFlinkApplication() {
	// We cannot test the success cases as they require a running CMF service. However we can test some basic failure cases.
	tests := []CLITest{
		// failure
		{args: "flink --url dummy-url application web-ui-forward forward-negative-port --environment forward-test --port -30", fixture: "flink/onprem/application/forward-negative-port.golden", exitCode: 1},
		{args: "flink --url dummy-url application web-ui-forward non-existent --environment default", fixture: "flink/onprem/application/forward-nonexistent-application.golden", exitCode: 1},
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
		{args: "flink environment use non-existent", fixture: "flink/onprem/environment/use-non-existent.golden", exitCode: 1},
		// success, followed by listing applications
		{args: "flink environment use default", fixture: "flink/onprem/environment/use-success.golden", env: []string{"HOME=" + tmpDir}},
		{args: "flink application describe default-application-1", fixture: "flink/onprem/application/describe-after-use-success.golden", env: []string{"HOME=" + tmpDir}},
	}

	for _, test := range tests {
		// Don't reset state between tests
		test.workflow = true
		s.runIntegrationTest(test)
	}
}
