package test

import (
	"os"

	"github.com/confluentinc/cli/v4/pkg/auth"
)

// Runs the integration test with login = "" and login = "onprem"
func runIntegrationTestsWithMultipleAuth(s *CLITestSuite, tests []CLITest) {
	for _, test := range tests {
		test.login = ""
		s.T().Setenv("LOGIN_TYPE", "")
		s.runIntegrationTest(test)

		test.name = test.args + "-onprem"
		test.login = "onprem"
		s.T().Setenv("LOGIN_TYPE", "onprem")
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkApplicationList() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink application list", fixture: "flink/application/list-env-missing.golden", exitCode: 1},
		{args: "flink application list --environment non-existent", fixture: "flink/application/list-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink application list --environment test", fixture: "flink/application/list-empty-env.golden"},
		{args: "flink application list --environment default  --output json", fixture: "flink/application/list-json.golden"},
		{args: "flink application list --environment default  --output human", fixture: "flink/application/list-human.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkApplicationDelete() {
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

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkEnvironmentList() {
	tests := []CLITest{
		// success scenarios
		{args: "flink environment list --output json", fixture: "flink/environment/list-json.golden"},
		{args: "flink environment list --output human", fixture: "flink/environment/list-human.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkEnvironmentDelete() {
	tests := []CLITest{
		// failure scenarios
		{args: "flink environment delete", fixture: "flink/environment/delete-env-missing.golden", exitCode: 1},
		{args: "flink environment delete non-existent", fixture: "flink/environment/delete-non-existent-env.golden", exitCode: 1},
		// success scenarios
		{args: "flink environment delete default test --force", fixture: "flink/environment/delete-success.golden"},
		// some failures and some successes
		{args: "flink environment delete default non-existent --force", fixture: "flink/environment/delete-mixed.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkApplicationCreate() {
	tests := []CLITest{
		// failure
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-unsuccessful-application.json", fixture: "flink/application/create-unsuccessful-application.golden", exitCode: 1},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-duplicate-application.json", fixture: "flink/application/create-duplicate-application.golden", exitCode: 1},
		{args: "flink application create --environment non-existent test/fixtures/input/flink/application/create-with-non-existent-environment.json", fixture: "flink/application/create-with-non-existent-environment.golden", exitCode: 1},
		// success
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json", fixture: "flink/application/create-success.golden"},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json --output yaml", fixture: "flink/application/create-success-yaml.golden"},
		// explicit test to see that even if the output is set to human, the output is still in json
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json --output human", fixture: "flink/application/create-with-human.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkApplicationUpdate() {
	tests := []CLITest{
		// failure
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-non-existent.json", fixture: "flink/application/update-non-existent.golden", exitCode: 1},
		{args: "flink application update --environment update-failure test/fixtures/input/flink/application/update-failure.json", fixture: "flink/application/update-failure.golden", exitCode: 1},
		{args: "flink application update --environment non-existent test/fixtures/input/flink/application/update-with-non-existent-environment.json", fixture: "flink/application/update-with-non-existent-environment.golden", exitCode: 1},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-with-get-failure.json", fixture: "flink/application/update-with-get-failure.golden", exitCode: 1},
		// success
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json", fixture: "flink/application/update-successful.golden"},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json --output yaml", fixture: "flink/application/update-successful-yaml.golden"},
		// explicit test to see that even if the output is set to human, the output is still in json
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json --output human", fixture: "flink/application/update-with-human.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkEnvironmentCreate() {
	tests := []CLITest{
		// success
		{args: "flink environment create default-2 --kubernetes-namespace default-staging", fixture: "flink/environment/create-success.golden"},
		{args: "flink environment create default-2 --kubernetes-namespace default-staging --output yaml", fixture: "flink/environment/create-success-yaml.golden"},
		{args: "flink environment create default-2 --defaults test/fixtures/input/flink/environment/application-defaults.json --kubernetes-namespace default-staging", fixture: "flink/environment/create-success-with-defaults.golden"},
		{args: "flink environment create default-2 --defaults test/fixtures/input/flink/environment/application-defaults.json --kubernetes-namespace default-staging --output json", fixture: "flink/environment/create-success-with-defaults-json.golden"},
		// failure
		{args: "flink environment create default-failure --kubernetes-namespace default-staging", fixture: "flink/environment/create-failure.golden", exitCode: 1},
		{args: "flink environment create default --kubernetes-namespace default-staging", fixture: "flink/environment/create-existing.golden", exitCode: 1},
		{args: "flink environment create default", fixture: "flink/environment/create-no-namespace.golden", exitCode: 1},
		// success with application, statement and compute pool defaults
		{args: "flink environment create default-2" +
			" --defaults test/fixtures/input/flink/environment/application-defaults.json" +
			" --statement-defaults test/fixtures/input/flink/environment/statement-defaults.json" +
			" --compute-pool-defaults test/fixtures/input/flink/environment/compute-pool-defaults.json" +
			" --kubernetes-namespace default-staging",
			fixture: "flink/environment/create-success-with-defaults-all.golden",
		},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkEnvironmentUpdate() {
	tests := []CLITest{
		// success
		{args: "flink environment update default --defaults '{\"property\": \"value\"}'", fixture: "flink/environment/update-success.golden"},
		{args: "flink environment update default --defaults '{\"property\": \"value\"}' --output yaml", fixture: "flink/environment/update-success-yaml.golden"},
		{args: "flink environment update default --defaults '{\"property\": \"value\"}' --output json", fixture: "flink/environment/update-success-json.golden"},
		// failure
		{args: "flink environment update update-failure --defaults '{\"property\": \"value\"}'", fixture: "flink/environment/update-failure.golden", exitCode: 1},
		{args: "flink environment update non-existent --defaults '{\"property\": \"value\"}'", fixture: "flink/environment/update-non-existent.golden", exitCode: 1},
		{args: "flink environment update get-failure --defaults '{\"property\": \"value\"}'", fixture: "flink/environment/update-get-failure.golden", exitCode: 1},
		{args: "flink environment update missing-flag-failure", fixture: "flink/environment/missing-flag-failure.golden", exitCode: 1},
		// success with application, statement and compute pool defaults
		{args: "flink environment update default" +
			" --defaults test/fixtures/input/flink/environment/application-defaults.json" +
			" --statement-defaults test/fixtures/input/flink/environment/statement-defaults.json" +
			" --compute-pool-defaults test/fixtures/input/flink/environment/compute-pool-defaults.json",
			fixture: "flink/environment/update-success-with-defaults-all.golden",
		},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkEnvironmentDescribe() {
	tests := []CLITest{
		// success
		{args: "flink environment describe default", fixture: "flink/environment/describe-success.golden"},
		{args: "flink environment describe default --output json", fixture: "flink/environment/describe-success-json.golden"},
		{args: "flink environment describe default --output yaml", fixture: "flink/environment/describe-success-yaml.golden"},
		{args: "flink environment describe defaults-all", fixture: "flink/environment/describe-success-with-defaults.golden"},
		// failure
		{args: "flink environment describe non-existent", fixture: "flink/environment/describe-non-existent.golden", exitCode: 1},
		{args: "flink environment describe", fixture: "flink/environment/describe-no-environment.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkApplicationDescribe() {
	tests := []CLITest{
		// success
		{args: "flink application describe --environment default default-application-1", fixture: "flink/application/describe-success.golden"},
		{args: "flink application describe --environment default default-application-1 --output yaml", fixture: "flink/application/describe-success-yaml.golden"},
		// explicit test to see that even if the output is set to human, the output is still in json
		{args: "flink application describe --environment default default-application-1 --output human", fixture: "flink/application/describe-with-human.golden"},
		// failure
		{args: "flink application describe --environment default non-existent", fixture: "flink/application/describe-non-existent.golden", exitCode: 1},
		{args: "flink application describe --environment non-existent default-application", fixture: "flink/application/describe-non-existent-environment.golden", exitCode: 1},
		{args: "flink application describe default-application-1", fixture: "flink/application/describe-no-environment.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkApplicationWebUiForward() {
	// We cannot test the success cases as they require a running CMF service. However we can test some basic failure cases.
	tests := []CLITest{
		// failure
		{args: "flink --url dummy-url application web-ui-forward forward-negative-port --environment forward-test --port -30", fixture: "flink/application/forward-negative-port.golden", exitCode: 1},
		{args: "flink --url dummy-url application web-ui-forward non-existent --environment default", fixture: "flink/application/forward-nonexistent-application.golden", exitCode: 1},
		{args: "flink --url dummy-url application web-ui-forward default-application-1 --environment non-existent", fixture: "flink/application/forward-nonexistent-environment.golden", exitCode: 1},
		{args: "flink --url dummy-url application web-ui-forward get-failure --environment default", fixture: "flink/application/forward-get-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)

	noUrlSetTest := CLITest{name: "no-url-set", args: "flink application web-ui-forward --environment does-not-matter missing-applications", fixture: "flink/application/url-missing.golden", exitCode: 1}
	// unset the environment variable
	os.Unsetenv(auth.ConfluentPlatformCmfURL)
	s.runIntegrationTest(noUrlSetTest)
}

func (s *CLITestSuite) TestFlinkComputePoolCreateOnPrem() {
	tests := []CLITest{
		// success
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.json --environment default", fixture: "flink/compute-pool/create-success.golden"},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.json --environment default --output yaml", fixture: "flink/compute-pool/create-success-yaml.golden"},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.json --environment default --output json", fixture: "flink/compute-pool/create-success-json.golden"},
		// failure
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-invalid-failure.json --environment default", fixture: "flink/compute-pool/create-invalid-failure.golden", exitCode: 1},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-existing-failure.json --environment default", fixture: "flink/compute-pool/create-existing-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkComputePoolDeleteOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink compute-pool delete test-pool1 --environment default", input: "y\n", fixture: "flink/compute-pool/delete-single-successful.golden"},
		{args: "flink compute-pool delete test-pool1 test-pool2 --environment default", input: "y\n", fixture: "flink/compute-pool/delete-multiple-successful.golden"},
		{args: "flink compute-pool delete test-pool1 --environment default --force", fixture: "flink/compute-pool/delete-single-force.golden"},
		// failure scenarios
		{args: "flink compute-pool delete test-pool1", fixture: "flink/compute-pool/delete-missing-env-flag-failure.golden", exitCode: 1},
		{args: "flink compute-pool delete test-pool1 --environment non-exist", fixture: "flink/compute-pool/delete-non-exist-env-failure.golden", exitCode: 1},
		{args: "flink compute-pool delete non-exist-pool --environment default", input: "y\n", fixture: "flink/compute-pool/delete-non-exist-pool-failure.golden", exitCode: 1},
		// mixed scenarios
		{args: "flink compute-pool delete test-pool1 non-exist-pool --environment default --force", fixture: "flink/compute-pool/delete-multiple-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkComputePoolDescribeOnPrem() {
	tests := []CLITest{
		// success
		{args: "flink compute-pool describe test-pool --environment default", fixture: "flink/compute-pool/describe-success.golden"},
		{args: "flink compute-pool describe test-pool --environment default --output yaml", fixture: "flink/compute-pool/describe-success-yaml.golden"},
		{args: "flink compute-pool describe test-pool --environment default --output json", fixture: "flink/compute-pool/describe-success-json.golden"},
		// failure
		{args: "flink compute-pool describe invalid-pool --environment default", fixture: "flink/compute-pool/describe-invalid-pool-failure.golden", exitCode: 1},
		{args: "flink compute-pool describe test-pool --environment non-exist", fixture: "flink/compute-pool/describe-non-exist-environment-failure.golden", exitCode: 1},
		{args: "flink compute-pool describe test-pool", fixture: "flink/compute-pool/describe-missing-env-flag-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkComputePoolListOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink compute-pool list --environment default", fixture: "flink/compute-pool/list-successful.golden"},
		{args: "flink compute-pool list --environment default  --output json", fixture: "flink/compute-pool/list-json.golden"},
		{args: "flink compute-pool list --environment default  --output human", fixture: "flink/compute-pool/list-human.golden"},
		// failure scenarios
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list-missing-env-flag-failure.golden", exitCode: 1},
		{args: "flink compute-pool list --environment non-exist", fixture: "flink/compute-pool/list-non-exist-environment-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkOnPremWithCloudLogin() {
	test := CLITest{args: "flink environment list --output json", fixture: "flink/environment/list-cloud.golden", login: "cloud", exitCode: 1}
	s.runIntegrationTest(test)
}
