package test

import (
	"os"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/go-prompt"

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
		{args: "flink application list --environment default  --output yaml", fixture: "flink/application/list-yaml.golden"},
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

func (s *CLITestSuite) TestFlinkSavepointCreate() {
	tests := []CLITest{
		{args: "flink savepoint create savepoint1 --environment default --application application1", fixture: "flink/savepoint/create-savepoint.golden"},
		{args: "flink savepoint create --environment default --application application2", fixture: "flink/savepoint/create-savepoint-no-name.golden"},
		{args: "flink savepoint create savepointS --environment default --statement test-stmt", fixture: "flink/savepoint/create-savepoint-statement.golden"},
		{args: "flink savepoint create savepointS --environment default --statement test-stmt --path abc/def --format NATIVE --backoff-limit 10", fixture: "flink/savepoint/create-savepoint-statement-values.golden"},
		// fail
		{args: "flink savepoint create savepoint1 --environment default --application application1 --statement statement1", fixture: "flink/savepoint/create-savepoint-fail-both.golden", exitCode: 1},
		{args: "flink savepoint create savepoint1 --environment default", fixture: "flink/savepoint/create-savepoint-fail-none.golden", exitCode: 1},
		{args: "flink savepoint create savepoint1 --application application1 --statement statement1", fixture: "flink/savepoint/create-savepoint-fail-no-env.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkSavepointListOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink savepoint list --environment default --application application1", fixture: "flink/savepoint/list-successful.golden"},
		{args: "flink savepoint list --environment default --statement statement1", fixture: "flink/savepoint/list-successful-statement.golden"},
		// failure scenarios
		{args: "flink savepoint list --environment default --statement statement1 --application application1", fixture: "flink/savepoint/list-fail-both.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkSavepointDescribe() {
	tests := []CLITest{
		{args: "flink savepoint describe savepoint1 --environment default --application application1", fixture: "flink/savepoint/describe-success.golden"},
		{args: "flink savepoint describe savepoint1 --environment default --statement statement1", fixture: "flink/savepoint/describe-success-statement.golden"},
		{args: "flink savepoint describe invalid-savepoint --environment default --statement statement1", fixture: "flink/savepoint/describe-fail.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkSavepointDeleteOnPrem() {
	tests := []CLITest{
		{args: "flink savepoint delete savepoint1 --environment default --application application1", input: "y\n", fixture: "flink/savepoint/delete-success.golden"},
		{args: "flink savepoint delete savepoint1 --environment default --statement statement1", input: "y\n", fixture: "flink/savepoint/delete-statement-success.golden"},
		{args: "flink savepoint delete savepoint1 --environment default --statement statement1 --force", fixture: "flink/savepoint/delete-force-success.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkDetachedSavepointCreate() {
	tests := []CLITest{
		{args: "flink detached-savepoint create savepoint1 --path abc/def", fixture: "flink/detached-savepoint/create-savepoint.golden"},
		{args: "flink detached-savepoint create savepoint1", fixture: "flink/detached-savepoint/create-savepoint-nopath.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkDetachedSavepointListOnPrem() {
	tests := []CLITest{
		{args: "flink detached-savepoint list", fixture: "flink/detached-savepoint/list-successful.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkDetachedSavepointDescribe() {
	tests := []CLITest{
		{args: "flink detached-savepoint describe savepoint1", fixture: "flink/detached-savepoint/describe-success.golden"},
		{args: "flink detached-savepoint describe invalid-savepoint", fixture: "flink/detached-savepoint/describe-fail.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkDetachedSavepointDeleteOnPrem() {
	tests := []CLITest{
		{args: "flink detached-savepoint delete savepoint1", input: "y\n", fixture: "flink/detached-savepoint/delete-success.golden"},
		{args: "flink detached-savepoint delete savepoint1 --force", fixture: "flink/detached-savepoint/delete-success-force.golden"},
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
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful-savepoint.json", fixture: "flink/application/update-successful-savepoint.golden"},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful-savepoint.json --output yaml", fixture: "flink/application/update-successful-savepoint-yaml.golden"},
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
		{args: "flink compute-pool list --environment default --output json", fixture: "flink/compute-pool/list-json.golden"},
		{args: "flink compute-pool list --environment default --output yaml", fixture: "flink/compute-pool/list-yaml.golden"},
		// failure scenarios
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list-missing-env-flag-failure.golden", exitCode: 1},
		{args: "flink compute-pool list --environment non-exist", fixture: "flink/compute-pool/list-non-exist-environment-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkCatalogCreateOnPrem() {
	tests := []CLITest{
		// success
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.json", fixture: "flink/catalog/create-success.golden"},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.json --output yaml", fixture: "flink/catalog/create-success-yaml.golden"},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.json --output json", fixture: "flink/catalog/create-success-json.golden"},
		// failure
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-invalid-failure.json", fixture: "flink/catalog/create-invalid-failure.golden", exitCode: 1},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-existing-failure.json", fixture: "flink/catalog/create-existing-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkCatalogDeleteOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink catalog delete test-catalog1", input: "y\n", fixture: "flink/catalog/delete-single-successful.golden"},
		{args: "flink catalog delete test-catalog1 test-catalog2", input: "y\n", fixture: "flink/catalog/delete-multiple-successful.golden"},
		{args: "flink catalog delete test-catalog1 --force", fixture: "flink/catalog/delete-single-force.golden"},
		// failure scenarios
		{args: "flink catalog delete non-exist-catalog", input: "y\n", fixture: "flink/catalog/delete-non-exist-catalog-failure.golden", exitCode: 1},
		// mixed scenarios
		{args: "flink catalog delete test-catalog1 non-exist-catalog --force", fixture: "flink/catalog/delete-multiple-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkCatalogDescribeOnPrem() {
	tests := []CLITest{
		// success
		{args: "flink catalog describe test-catalog1", fixture: "flink/catalog/describe-success.golden"},
		{args: "flink catalog describe test-catalog1 --output yaml", fixture: "flink/catalog/describe-success-yaml.golden"},
		{args: "flink catalog describe test-catalog1 --output json", fixture: "flink/catalog/describe-success-json.golden"},
		// failure
		{args: "flink catalog describe invalid-catalog", fixture: "flink/catalog/describe-non-exist-catalog-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkCatalogListOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink catalog list", fixture: "flink/catalog/list-successful.golden"},
		{args: "flink catalog list --output json", fixture: "flink/catalog/list-json.golden"},
		{args: "flink catalog list --output yaml", fixture: "flink/catalog/list-yaml.golden"},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkStatementCreateOnPrem() {
	tests := []CLITest{
		// success
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool`, fixture: "flink/statement/create-success.golden"},
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool -o json`, fixture: "flink/statement/create-success-json.golden"},
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool -o yaml`, fixture: "flink/statement/create-success-yaml.golden"},
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool --flink-configuration test/fixtures/input/flink/statement/flink-configuration.json`, fixture: "flink/statement/create-success.golden"},
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool --flink-configuration test/fixtures/input/flink/statement/flink-configuration.yaml`, fixture: "flink/statement/create-success.golden"},
		// failure
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool --flink-configuration test/fixtures/input/flink/statement/flink-configuration.properties`, fixture: "flink/statement/create-failure-invalid-configuration-file-format.golden", exitCode: 1},
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool --flink-configuration test/fixtures/input/flink/statement/flink-configuration.csv`, fixture: "flink/statement/create-failure-configuration-file-dne.golden", regex: true, exitCode: 1},
		{args: "flink statement create test-stmt --environment default --compute-pool test-pool", fixture: "flink/statement/create-missing-sql-failure.golden", exitCode: 1},
		{args: `flink statement create test-stmt --environment default --sql "SELECT * FROM test_table"`, fixture: "flink/statement/create-missing-compute-pool-failure.golden", exitCode: 1},
		{args: `flink statement create invalid-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool`, fixture: "flink/statement/create-invalid-stmt-failure.golden", exitCode: 1},
		{args: `flink statement create existing-stmt --environment default --sql "SELECT * FROM test_table" --compute-pool test-pool`, fixture: "flink/statement/create-existing-stmt-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkStatementDescribeOnPrem() {
	tests := []CLITest{
		// success
		{args: "flink statement describe test-stmt --environment default", fixture: "flink/statement/describe-success.golden"},
		{args: "flink statement describe test-stmt --environment default -o json", fixture: "flink/statement/describe-success-json.golden"},
		{args: "flink statement describe test-stmt --environment default -o yaml", fixture: "flink/statement/describe-success-yaml.golden"},
		{args: "flink statement describe shell-test-stmt --environment default -o json", fixture: "flink/statement/describe-success-completed-json.golden"},
		// failure
		{args: "flink statement describe test-stmt", fixture: "flink/statement/describe-env-missing-failure.golden", exitCode: 1},
		{args: "flink statement describe test-stmt --environment non-exist", fixture: "flink/statement/describe-non-exist-env-failure.golden", exitCode: 1},
		{args: "flink statement describe invalid-stmt --environment default", fixture: "flink/statement/describe-invalid-stmt-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkStatementDeleteOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink statement delete test-stmt1 --environment default", input: "y\n", fixture: "flink/statement/delete-single-successful.golden"},
		{args: "flink statement delete test-stmt1 test-stmt2 --environment default", input: "y\n", fixture: "flink/statement/delete-multiple-successful.golden"},
		{args: "flink statement delete test-stmt1 --environment default --force", fixture: "flink/statement/delete-single-force.golden"},
		// failure scenarios
		{args: "flink statement delete non-exist-stmt --environment default", input: "y\n", fixture: "flink/statement/delete-non-exist-statement-failure.golden", exitCode: 1},
		{args: "flink statement delete test-stmt1 --environment non-exist", input: "y\n", fixture: "flink/statement/delete-non-exist-env-failure.golden", exitCode: 1},
		{args: "flink statement delete test-stmt1", input: "y\n", fixture: "flink/statement/delete-missing-env-failure.golden", exitCode: 1},
		// mixed scenarios
		{args: "flink statement delete test-stmt1 non-exist-stmt --environment default --force", fixture: "flink/statement/delete-multiple-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkStatementListOnPrem() {
	tests := []CLITest{
		// success
		{args: "flink statement list --environment default", fixture: "flink/statement/list-success.golden"},
		{args: "flink statement list --environment default -o json", fixture: "flink/statement/list-success-json.golden"},
		{args: "flink statement list --environment default -o yaml", fixture: "flink/statement/list-success-yaml.golden"},
		// failure
		{args: "flink statement list", fixture: "flink/statement/list-env-missing-failure.golden", exitCode: 1},
		{args: "flink statement list --environment non-exist", fixture: "flink/statement/list-non-exist-env-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkStatementUpdateOnPrem() {
	tests := []CLITest{
		// stop scenarios:
		{args: "flink statement stop test-stmt1 --environment default", fixture: "flink/statement/stop-successful.golden"},
		{args: "flink statement stop non-exist-stmt --environment default", fixture: "flink/statement/stop-failure.golden", exitCode: 1},
		{args: "flink statement stop test-stmt1", fixture: "flink/statement/stop-missing-env-failure.golden", exitCode: 1},

		// resume scenarios:
		{args: "flink statement resume test-stmt1 --environment default", fixture: "flink/statement/resume-successful.golden"},
		{args: "flink statement resume non-exist-stmt --environment default", fixture: "flink/statement/resume-failure.golden", exitCode: 1},
		{args: "flink statement resume test-stmt1", fixture: "flink/statement/resume-missing-env-failure.golden", exitCode: 1},

		// rescale scenarios:
		{args: "flink statement rescale test-stmt1 --parallelism 4 --environment default", fixture: "flink/statement/rescale-successful.golden"},
		{args: "flink statement rescale non-exist-stmt --environment default", fixture: "flink/statement/rescale-failure.golden", exitCode: 1},
		{args: "flink statement rescale test-stmt1 --parallelism 4", fixture: "flink/statement/rescale-missing-env-failure.golden", exitCode: 1},
		{args: "flink statement rescale test-stmt1 --environment default", fixture: "flink/statement/rescale-missing-parallelism-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkStatementExceptionListOnPrem() {
	tests := []CLITest{
		// success scenarios
		{args: "flink statement exception list test-stmt1 --environment default", fixture: "flink/statement/list-exceptions-successful.golden"},
		// failure scenarios
		{args: "flink statement exception list test-stmt1 --environment non-exist", fixture: "flink/statement/list-exceptions-non-exist-env-failure.golden", exitCode: 1},
		{args: "flink statement exception list invalid-stmt --environment default", fixture: "flink/statement/list-exceptions-invalid-stmt-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkOnPremWithCloudLogin() {
	test := CLITest{args: "flink environment list --output json", fixture: "flink/environment/list-cloud.golden", login: "cloud", exitCode: 1}
	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestFlinkApplicationCreateWithYAML() {
	tests := []CLITest{
		// failure scenarios with YAML files
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-unsuccessful-application.json", fixture: "flink/application/create-unsuccessful-application.golden", exitCode: 1},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-duplicate-application.json", fixture: "flink/application/create-duplicate-application.golden", exitCode: 1},
		{args: "flink application create --environment non-existent test/fixtures/input/flink/application/create-with-non-existent-environment.json", fixture: "flink/application/create-with-non-existent-environment.golden", exitCode: 1},
		// success scenarios with YAML files
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json", fixture: "flink/application/create-success.golden"},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json --output yaml", fixture: "flink/application/create-success-yaml.golden"},
		// explicit test to see that even if the output is set to human, the output is still in json
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.json --output human", fixture: "flink/application/create-with-human.golden"},
		// YAML file tests
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.yaml", fixture: "flink/application/create-success.golden"},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.yaml --output yaml", fixture: "flink/application/create-success-yaml.golden"},
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-new.yaml --output human", fixture: "flink/application/create-with-human.golden"},
		// YAML file failure scenarios
		{args: "flink application create --environment default test/fixtures/input/flink/application/create-duplicate-application.yaml", fixture: "flink/application/create-duplicate-application.golden", exitCode: 1},
		{args: "flink application create --environment non-existent test/fixtures/input/flink/application/create-with-non-existent-environment.yaml", fixture: "flink/application/create-with-non-existent-environment.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkApplicationUpdateWithYAML() {
	tests := []CLITest{
		// failure scenarios with JSON files
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-non-existent.json", fixture: "flink/application/update-non-existent.golden", exitCode: 1},
		{args: "flink application update --environment update-failure test/fixtures/input/flink/application/update-failure.json", fixture: "flink/application/update-failure.golden", exitCode: 1},
		{args: "flink application update --environment non-existent test/fixtures/input/flink/application/update-with-non-existent-environment.json", fixture: "flink/application/update-with-non-existent-environment.golden", exitCode: 1},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-with-get-failure.json", fixture: "flink/application/update-with-get-failure.golden", exitCode: 1},
		// success scenarios with JSON files
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json", fixture: "flink/application/update-successful.golden"},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json --output yaml", fixture: "flink/application/update-successful-yaml.golden"},
		// explicit test to see that even if the output is set to human, the output is still in json
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.json --output human", fixture: "flink/application/update-with-human.golden"},
		// YAML file tests
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.yaml", fixture: "flink/application/update-successful.golden"},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.yaml --output yaml", fixture: "flink/application/update-successful-yaml.golden"},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-successful.yaml --output human", fixture: "flink/application/update-with-human.golden"},
		// YAML file failure scenarios
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-non-existent.yaml", fixture: "flink/application/update-non-existent.golden", exitCode: 1},
		{args: "flink application update --environment update-failure test/fixtures/input/flink/application/update-failure.json", fixture: "flink/application/update-failure.golden", exitCode: 1},
		{args: "flink application update --environment non-existent test/fixtures/input/flink/application/update-with-non-existent-environment.json", fixture: "flink/application/update-with-non-existent-environment.golden", exitCode: 1},
		{args: "flink application update --environment default test/fixtures/input/flink/application/update-with-get-failure.json", fixture: "flink/application/update-with-get-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkComputePoolCreateWithYAML() {
	tests := []CLITest{
		// success scenarios with JSON files
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.json --environment default", fixture: "flink/compute-pool/create-success.golden"},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.json --environment default --output yaml", fixture: "flink/compute-pool/create-success-yaml.golden"},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.json --environment default --output json", fixture: "flink/compute-pool/create-success-json.golden"},
		// failure scenarios with JSON files
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-invalid-failure.json --environment default", fixture: "flink/compute-pool/create-invalid-failure.golden", exitCode: 1},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-existing-failure.json --environment default", fixture: "flink/compute-pool/create-existing-failure.golden", exitCode: 1},
		// YAML file tests
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.yaml --environment default", fixture: "flink/compute-pool/create-success.golden"},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.yaml --environment default --output yaml", fixture: "flink/compute-pool/create-success-yaml.golden"},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-successful.yaml --environment default --output json", fixture: "flink/compute-pool/create-success-json.golden"},
		// YAML file failure scenarios
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-invalid-failure.yaml --environment default", fixture: "flink/compute-pool/create-invalid-failure.golden", exitCode: 1},
		{args: "flink compute-pool create test/fixtures/input/flink/compute-pool/create-existing-failure.yaml --environment default", fixture: "flink/compute-pool/create-existing-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkCatalogCreateWithYAML() {
	tests := []CLITest{
		// success scenarios with JSON files
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.json", fixture: "flink/catalog/create-success.golden"},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.json --output yaml", fixture: "flink/catalog/create-success-yaml.golden"},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.json --output json", fixture: "flink/catalog/create-success-json.golden"},
		// failure scenarios with JSON files
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-invalid-failure.json", fixture: "flink/catalog/create-invalid-failure.golden", exitCode: 1},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-existing-failure.json", fixture: "flink/catalog/create-existing-failure.golden", exitCode: 1},
		// YAML file tests
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.yaml", fixture: "flink/catalog/create-success.golden"},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.yaml --output yaml", fixture: "flink/catalog/create-success-yaml.golden"},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-successful.yaml --output json", fixture: "flink/catalog/create-success-json.golden"},
		// YAML file failure scenarios
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-invalid-failure.yaml", fixture: "flink/catalog/create-invalid-failure.golden", exitCode: 1},
		{args: "flink catalog create test/fixtures/input/flink/catalog/create-existing-failure.yaml", fixture: "flink/catalog/create-existing-failure.golden", exitCode: 1},
	}

	runIntegrationTestsWithMultipleAuth(s, tests)
}

func (s *CLITestSuite) TestFlinkShellOnPrem() {
	tests := []flinkShellTest{
		{
			goldenFile: "use-catalog-onprem.golden",
			commands: []string{
				"use catalog default;",
				"set;",
			},
		},
		{
			goldenFile: "use-database-onprem.golden",
			commands: []string{
				"use catalog default;",
				"use db1;",
				"set;",
			},
		},
		{
			goldenFile: "set-single-key-onprem.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"set;",
			},
		},
		{
			goldenFile: "reset-single-key-onprem.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"reset 'cli.a-key';",
				"set;",
			},
		},
		{
			goldenFile: "reset-all-keys-onprem.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"set 'cli.another-key'='another value';",
				"reset;",
				"set;",
			},
		},
		{
			goldenFile: "shell-describe-table-onprem.golden",
			commands: []string{
				"set 'client.statement-name'='shell-test-stmt';",
				"describe clicks;",
				"set;",
			},
		},
	}

	s.setupFlinkShellTestsOnPrem()
	defer s.tearDownFlinkShellTests()

	for _, test := range tests {
		test.isOnPrem = true
		s.runFlinkShellTest(test)
	}
	/*s.loginOnPrem(s.T())
	for _, test := range tests {
		test.isOnPrem = true
		s.runFlinkShellTest(test)
	}*/

	resetConfiguration(s.T(), false)
}

/*func (s *CLITestSuite) loginOnPrem(t *testing.T) {
	loginString := fmt.Sprintf("login --url %s", s.TestBackend.GetMdsUrl())
	env := []string{pauth.ConfluentPlatformUsername + "=fake@user.com", pauth.ConfluentPlatformPassword + "=pass1"}
	if output := runCommand(t, testBin, env, loginString, 0, ""); *debug {
		fmt.Println(output)
	}
}*/

func (s *CLITestSuite) setupFlinkShellTestsOnPrem() {
	// Set the go-prompt file input env var, so go-prompt uses this file as the input stream
	err := os.Setenv(prompt.EnvVarInputFile, flinkShellInputStreamFile)
	require.NoError(s.T(), err)

	// Fake the timezone, to ensure CI and local run with the same default timezone.
	// We use UTC to avoid time zone differences due to daylight savings time.
	err = os.Setenv(timezoneEnvVar, "UTC")
	require.NoError(s.T(), err)
}
