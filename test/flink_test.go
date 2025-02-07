package test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/go-prompt"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
)

const (
	flinkShellInputStreamFile     = "flink_shell_input_stream.txt"
	flinkShellFixtureOutputFolder = "test/fixtures/output/flink/shell"
	timezoneEnvVar                = "TZ"
	flinkShellTimeout             = 10 * time.Second
)

type flinkShellTest struct {
	commands   []string
	goldenFile string
}

func (s *CLITestSuite) TestFlinkArtifact() {
	tests := []CLITest{
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/java-udf-examples-3.0.jar --cloud aws --region us-west-2 --environment env-123456", fixture: "flink/artifact/create.golden"},
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/java-udf-examples-3.0.jar --cloud aws --region us-west-2 --environment env-123456 --description CliArtifactTest", fixture: "flink/artifact/create.golden"},
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/java-udf-examples-3.0.jar --cloud aws --region us-west-2 --environment env-123456 --description CliArtifactTest --documentation-link https://docs.confluent.io", fixture: "flink/artifact/create.golden"},
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/python-udf-examples.zip --cloud aws --region us-west-2 --environment env-789012 --description CliArtifactTest --runtime-language python", fixture: "flink/artifact/create-python.golden"},
		{args: "flink artifact describe --cloud aws --region us-west-2 --environment env-123456 cfa-789013", fixture: "flink/artifact/describe.golden"},
		{args: "flink artifact list --cloud aws --region us-west-2 --environment env-123456", fixture: "flink/artifact/list.golden"},
		{args: "flink artifact delete --cloud aws --region us-west-2 --environment env-123456 --force cfa-123456", fixture: "flink/artifact/delete.golden"},
		{args: "flink artifact delete --cloud aws --region us-west-2 --environment env-123456 cfa-123456", input: "y\n", fixture: "flink/artifact/delete-prompt.golden"},

		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/java-udf-examples-3.0.jar --cloud azure --region centralus --environment env-123456", fixture: "flink/artifact/create-azure.golden"},
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/java-udf-examples-3.0.jar --cloud azure --region centralus --environment env-123456 --description CliArtifactTest", fixture: "flink/artifact/create-azure.golden"},
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/java-udf-examples-3.0.jar --cloud azure --region centralus --environment env-123456 --description CliArtifactTest --documentation-link https://docs.confluent.io", fixture: "flink/artifact/create-azure.golden"},
		{args: "flink artifact create my-flink-artifact --artifact-file test/fixtures/input/flink/python-udf-examples.zip --cloud azure --region centralus --environment env-789012 --description CliArtifactTest --runtime-language python", fixture: "flink/artifact/create-python-azure.golden"},
		{args: "flink artifact describe --cloud azure --region centralus --environment env-123456 cfa-789013", fixture: "flink/artifact/describe-azure.golden"},
		{args: "flink artifact list --cloud azure --region centralus --environment env-123456", fixture: "flink/artifact/list-azure.golden"},
		{args: "flink artifact delete --cloud azure --region centralus --environment env-123456 --force cfa-123456", fixture: "flink/artifact/delete-azure.golden"},
		{args: "flink artifact delete --cloud azure --region centralus --environment env-123456 cfa-123456", input: "y\n", fixture: "flink/artifact/delete-prompt-azure.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePool() {
	tests := []CLITest{
		{args: "flink compute-pool create my-compute-pool --cloud aws --region us-west-2", fixture: "flink/compute-pool/create.golden"},
		{args: "flink compute-pool describe lfcp-123456", fixture: "flink/compute-pool/describe.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list.golden"},
		{args: "flink compute-pool list --region eu-west-2", fixture: "flink/compute-pool/list-region.golden"},
		{args: "flink compute-pool update lfcp-123456 --max-cfu 5", fixture: "flink/compute-pool/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkConnection() {
	tests := []CLITest{
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type openai --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create.golden"},
		{args: "flink connection describe my-connection --cloud aws --region eu-west-1", fixture: "flink/connection/describe/describe.golden"},
		{args: "flink connection describe nonexist-connection --cloud aws --region eu-west-1", fixture: "flink/connection/describe/describe-nonexist.golden", exitCode: 1},
		{args: "flink connection list --cloud aws --region eu-west-1", fixture: "flink/connection/list/list.golden"},
		{args: "flink connection list --type openai --cloud aws --region eu-west-1", fixture: "flink/connection/list/list.golden"},
		{args: "flink connection list --type invalid --cloud aws --region eu-west-1", fixture: "flink/connection/list/list-failure.golden", exitCode: 1},
		{args: "flink connection update my-connection --cloud aws --region eu-west-1 --api-key 0000000000000000", fixture: "flink/connection/update/update.golden"},
		{args: "flink connection update nonexist-connection --cloud aws --region eu-west-1 --api-key 0000000000000000", fixture: "flink/connection/update/update-nonexist.golden", exitCode: 1},
		{args: "flink connection update my-connection --cloud aws --region eu-west-1 --service-key 0000000000000000", fixture: "flink/connection/update/update-wrong-secret.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkConnectionWrongEnv() {
	tests := []CLITest{
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type openai --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000 --environment env-dne", fixture: "flink/connection/create/create-wrong-env.golden", exitCode: 1},
		{args: "flink connection describe my-connection --cloud aws --region eu-west-1 --environment env-dne", fixture: "flink/connection/describe/describe-wrong-env.golden", exitCode: 1},
		{args: "flink connection list --cloud aws --region eu-west-1 --environment env-dne", fixture: "flink/connection/list/list-wrong-env.golden", exitCode: 1},
		{args: "flink connection update my-connection --cloud aws --region eu-west-1 --api-key 0000000000000000 --environment env-dne", fixture: "flink/connection/update/update-wrong-env.golden", exitCode: 1},
		{args: "flink connection delete my-connection --force --cloud aws --region eu-west-1 --environment env-dne", fixture: "flink/connection/delete/delete-wrong-env.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkConnectionDelete() {
	tests := []CLITest{
		{args: "flink connection delete my-connection --force --cloud aws --region eu-west-1", fixture: "flink/connection/delete/delete.golden"},
		{args: "flink connection delete my-connection my-connection-1 --cloud aws --region eu-west-1", input: "n\n", fixture: "flink/connection/delete/delete-multiple-refuse.golden"},
		{args: "flink connection delete my-connection my-connection-1 --cloud aws --region eu-west-1", input: "y\n", fixture: "flink/connection/delete/delete-multiple-success.golden"},
		{args: "flink connection delete my-connection nonexist-connection --cloud aws --region eu-west-1", fixture: "flink/connection/delete/delete-multiple-failure.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkConnectionCreateFailure() {
	tests := []CLITest{
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type OPENAI --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create-wrong-type.golden", exitCode: 1},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type openai --endpoint https://api.openai.com/v1/chat/completions --username username", fixture: "flink/connection/create/create-wrong-secret-type.golden", exitCode: 1},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type bedrock --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000 --aws-access-key 0000000000000000 --aws-secret-key 0000000000000000 --aws-session-token 0000000000000000", fixture: "flink/connection/create/create-extra-secret.golden", exitCode: 1},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type bedrock --endpoint https://api.openai.com/v1/chat/completions --aws-secret-key 0000000000000000 --aws-session-token 0000000000000000", fixture: "flink/connection/create/create-missing-required-secret.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkConnectionCreateSuccess() {
	tests := []CLITest{
		{args: "flink connection create my-connection --cloud azure --region eu-west-1 --type azureml --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create-azure.golden"},
		{args: "flink connection create my-connection --cloud azure --region eu-west-1 --type azureopenai --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create-azureopenai.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type bedrock --endpoint https://api.openai.com/v1/chat/completions --aws-access-key 0000000000000000 --aws-secret-key 0000000000000000 --aws-session-token 0000000000000000", fixture: "flink/connection/create/create-bedrock.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type bedrock --endpoint https://api.openai.com/v1/chat/completions --aws-access-key 0000000000000000 --aws-secret-key 0000000000000000", fixture: "flink/connection/create/create-bedrock.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type sagemaker --endpoint https://api.openai.com/v1/chat/completions --aws-access-key 0000000000000000 --aws-secret-key 0000000000000000 --aws-session-token 0000000000000000", fixture: "flink/connection/create/create-sagemaker.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type sagemaker --endpoint https://api.openai.com/v1/chat/completions --aws-access-key 0000000000000000 --aws-secret-key 0000000000000000", fixture: "flink/connection/create/create-sagemaker.golden"},
		{args: "flink connection create my-connection --cloud gcp --region eu-west-1 --type googleai --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create-googleai.golden"},
		{args: "flink connection create my-connection --cloud gcp --region eu-west-1 --type vertexai --endpoint https://api.openai.com/v1/chat/completions --service-key 0000000000000000", fixture: "flink/connection/create/create-vertexai.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type mongodb --endpoint https://api.openai.com/v1/chat/completions --username name --password pass", fixture: "flink/connection/create/create-mongodb.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type elastic --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create-elastic.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type pinecone --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000", fixture: "flink/connection/create/create-pinecone.golden"},
		{args: "flink connection create my-connection --cloud aws --region eu-west-1 --type couchbase --endpoint https://api.openai.com/v1/chat/completions --username name --password pass", fixture: "flink/connection/create/create-couchbase.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkConnectivityType() {
	listPrivateFixture := "flink/connectivity-type/list-private.golden"
	if runtime.GOOS == "windows" { // Error message is different on Windows
		listPrivateFixture = "flink/connectivity-type/list-private-windows.golden"
	}

	tests := []CLITest{
		{args: "flink connectivity-type use public", fixture: "flink/connectivity-type/use-public.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1", fixture: "flink/connectivity-type/list-public.golden"},
		{args: "flink connectivity-type use private", fixture: "flink/connectivity-type/use-private.golden"},
		// Checking that the private endpoint is getting hit. The error here tells us that we are not using the public URL anymore.
		{args: "flink statement list --cloud aws --region eu-west-1", fixture: listPrivateFixture, exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolDelete() {
	tests := []CLITest{
		{args: "flink compute-pool delete lfcp-123456 --force", fixture: "flink/compute-pool/delete.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-222222", input: "n\n", fixture: "flink/compute-pool/delete-multiple-refuse.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-222222", input: "y\n", fixture: "flink/compute-pool/delete-multiple-success.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-654321", fixture: "flink/compute-pool/delete-multiple-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUnset() {
	tests := []CLITest{
		{args: "flink compute-pool unset", login: "cloud", fixture: "flink/compute-pool/unset-before-use.golden"},
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use-before-unset.golden"},
		{args: "flink compute-pool unset", login: "cloud", fixture: "flink/compute-pool/unset.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUse() {
	tests := []CLITest{
		{args: "flink compute-pool use lfcp-999999", login: "cloud", fixture: "flink/compute-pool/use-fail.golden", exitCode: 1},
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use.golden"},
		{args: "flink compute-pool describe", fixture: "flink/compute-pool/describe-after-use.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list-after-use.golden"},
		{args: "flink compute-pool update --max-cfu 5", fixture: "flink/compute-pool/update-after-use.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkRegion() {
	tests := []CLITest{
		{args: "flink region use --cloud aws --region eu-west-1", fixture: "flink/region/use.golden"},
		{args: "flink region list", fixture: "flink/region/list.golden"},
		{args: "flink region list -o json", fixture: "flink/region/list-json.golden"},
		{args: "flink region list --cloud aws", fixture: "flink/region/list-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatement() {
	tests := []CLITest{
		{args: "flink statement delete my-statement --force --cloud aws --region eu-west-1", fixture: "flink/statement/delete.golden"},
		{args: "flink statement describe my-statement --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/describe-yaml.golden"},
		{args: "flink statement describe my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/describe.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1", fixture: "flink/statement/list.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/list-yaml.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --status completed", fixture: "flink/statement/list-completed.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --status pending", fixture: "flink/statement/list-pending.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --compute-pool lfcp-nonexistent", fixture: "flink/statement/list-cp-not-found.golden", exitCode: 1},
		{args: "flink statement list --cloud aws --region eu-west-2 --compute-pool lfcp-123456", fixture: "flink/statement/list-cp-incorrect-region.golden", exitCode: 1},
		{args: "flink statement stop my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/stop.golden"},
		{args: "flink statement resume my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/resume-valid.golden"},
		{args: "flink statement resume my-statement --cloud aws --region eu-west-1 --principal u-123456", fixture: "flink/statement/resume-valid.golden"},
		{args: "flink statement resume my-statement --cloud aws --region eu-west-1 --compute-pool lfcp-123456", fixture: "flink/statement/resume-valid.golden"},
		{args: "flink statement resume my-statement --cloud aws --region eu-west-1 --principal u-123456 --compute-pool lfcp-123456", fixture: "flink/statement/resume-valid.golden"},
		{args: "flink statement resume my-statement --cloud aws --region eu-west-1 --principal sa-654321", fixture: "flink/statement/resume-invalid-principal.golden", exitCode: 1},
		{args: "flink statement resume my-statement --cloud aws --region eu-west-1 --compute-pool lfcp-654321", fixture: "flink/statement/resume-invalid-compute-pool.golden", exitCode: 1},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --compute-pool lfcp-123456", fixture: "flink/statement/update-compute-pool.golden"},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --principal u-123456", fixture: "flink/statement/update-principal.golden"},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --stopped=false", fixture: "flink/statement/update-stopped.golden"},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --stopped=false --principal u-123456", fixture: "flink/statement/update-stopped.golden"},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --stopped=false --compute-pool lfcp-123456", fixture: "flink/statement/update-stopped.golden"},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --stopped=false --compute-pool lfcp-123456 --principal u-123456", fixture: "flink/statement/update-stopped.golden"},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --stopped=false --compute-pool lfcp-654321", fixture: "flink/statement/update-invalid-compute-pool.golden", exitCode: 1},
		{args: "flink statement update my-statement --cloud aws --region eu-west-1 --stopped=false --principal u-654321", fixture: "flink/statement/update-invalid-principal.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatementCreate() {
	tests := []CLITest{
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "flink/statement/create.golden"},
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456`, fixture: "flink/statement/create-service-account-warning.golden"},
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456 --wait`, fixture: "flink/statement/create-wait.golden"},
		{args: `flink statement create --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456 -o yaml`, fixture: "flink/statement/create-no-name-yaml.golden", regex: true},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlink_Autocomplete() {
	tests := []CLITest{
		{args: `__complete flink compute-pool create my-compute-pool --cloud ""`, fixture: "flink/compute-pool/create-cloud-autocomplete.golden"},
		{args: `__complete flink compute-pool create my-compute-pool --cloud aws --region ""`, fixture: "flink/compute-pool/create-region-autocomplete.golden"},
		{args: `__complete flink compute-pool delete ""`, fixture: "flink/compute-pool/delete-autocomplete.golden"},
		{args: `__complete flink compute-pool list --region ""`, fixture: "flink/compute-pool/list-region-autocomplete.golden"},
		{args: `__complete flink statement create my-statement --database ""`, fixture: "flink/statement/create-database-autocomplete.golden"},
		{args: `__complete flink statement update my-statement --principal ""`, fixture: "flink/statement/update-database-autocomplete.golden"},
		{args: `__complete flink connection create my-connection --cloud ""`, fixture: "flink/connection/create/create-cloud-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkShell() {
	tests := []flinkShellTest{
		{
			goldenFile: "use-catalog.golden",
			commands: []string{
				"use catalog default;",
				"set;",
			},
		},
		{
			goldenFile: "use-database.golden",
			commands: []string{
				"use db1;",
				"set;",
			},
		},
		{
			goldenFile: "set-single-key.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"set;",
			},
		},
		{
			goldenFile: "reset-single-key.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"reset 'cli.a-key';",
				"set;",
			},
		},
		{
			goldenFile: "reset-all-keys.golden",
			commands: []string{
				"set 'cli.a-key'='a value';",
				"set 'cli.another-key'='another value';",
				"reset;",
				"set;",
			},
		},
	}

	s.setupFlinkShellTests()
	defer s.tearDownFlinkShellTests()
	for _, test := range tests {
		s.runFlinkShellTest(test)
	}
}

func (s *CLITestSuite) setupFlinkShellTests() {
	s.login(s.T())

	// Set the go-prompt file input env var, so go-prompt uses this file as the input stream
	err := os.Setenv(prompt.EnvVarInputFile, flinkShellInputStreamFile)
	require.NoError(s.T(), err)

	// Fake the timezone, to ensure CI and local run with the same default timezone.
	// We use UTC to avoid time zone differences due to daylight savings time.
	err = os.Setenv(timezoneEnvVar, "UTC")
	require.NoError(s.T(), err)
}

func (s *CLITestSuite) login(t *testing.T) {
	loginString := fmt.Sprintf("login --url %s", s.TestBackend.GetCloudUrl())
	env := []string{pauth.ConfluentCloudEmail + "=fake@user.com", pauth.ConfluentCloudPassword + "=pass1"}
	if output := runCommand(t, testBin, env, loginString, 0, ""); *debug {
		fmt.Println(output)
	}
}

func (s *CLITestSuite) tearDownFlinkShellTests() {
	err := os.Unsetenv(prompt.EnvVarInputFile)
	require.NoError(s.T(), err)

	err = os.Unsetenv(timezoneEnvVar)
	require.NoError(s.T(), err)
}

func (s *CLITestSuite) runFlinkShellTest(flinkShellTest flinkShellTest) {
	testName := strings.TrimSuffix(flinkShellTest.goldenFile, ".golden")
	s.T().Run(testName, func(t *testing.T) {
		// Create a file for go-prompt to use as the input stream
		stdin, err := os.Create(flinkShellInputStreamFile)
		require.NoError(s.T(), err, "error creating file")
		defer func() {
			require.NoError(t, cleanupInputFile(stdin))
		}()

		// Start flink shell
		dir, err := os.Getwd()
		require.NoError(t, err)
		cmd := exec.Command(filepath.Join(dir, testBin), "flink", "shell", "--compute-pool", "lfcp-123456")

		// Register stdout scanner
		pipe, err := cmd.StdoutPipe()
		require.NoError(t, err)
		stdoutScanner := bufio.NewScanner(pipe)

		// Start command
		err = cmd.Start()
		require.NoError(t, err)

		output := &strings.Builder{}
		output.WriteString(waitForLine(stdoutScanner, "[Ctrl-Q] Quit [Ctrl-S] Toggle Completions"))

		// Execute commands
		require.NoError(t, err)
		outputFromCommands, err := executeCommands(stdin, flinkShellTest.commands, stdoutScanner)
		require.NoError(t, err)
		output.WriteString(outputFromCommands)

		cmdDone := make(chan error)
		go func() {
			cmdDone <- cmd.Wait()
		}()

		// Wait for flink shell to exit or timeout
		select {
		case err := <-cmdDone:
			require.NoError(t, err)
		case <-time.After(flinkShellTimeout):
			require.NoError(t, cmd.Process.Kill())
			require.NoError(t, pipe.Close())
			t.Fatalf("test timed out")
		}

		// Compare to golden file
		snapshotConfig := cupaloy.New(
			cupaloy.SnapshotSubdirectory(filepath.Join(dir, flinkShellFixtureOutputFolder)),
			// Update snapshot if update flag was set
			cupaloy.ShouldUpdate(func() bool {
				return *update
			}),
		)
		assert.NoError(t, snapshotConfig.SnapshotWithName(flinkShellTest.goldenFile, output.String()),
			fmt.Sprintf("full output was %s", output.String()))
	})
}

func cleanupInputFile(file *os.File) error {
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Remove(file.Name()); err != nil {
		return err
	}
	return nil
}

func executeCommands(stdin *os.File, commands []string, stdoutScanner *bufio.Scanner) (string, error) {
	// add exit command to ensure we always close the flink shell
	commands = append(commands, "exit")
	output := strings.Builder{}
	for _, command := range commands {
		// Simulate the user entering a command and add a new line to flush the output buffer
		_, err := stdin.WriteString(command + "\n")
		if err != nil {
			return "", err
		}

		output.WriteString(waitForLine(stdoutScanner, fmt.Sprintf("> %s", command)))

		// submit the statement
		_, err = stdin.WriteString("\n")
		if err != nil {
			return "", err
		}

		output.WriteString(waitForLine(stdoutScanner, "Statement successfully submitted."))
	}
	return output.String(), nil
}

func waitForLine(stdoutScanner *bufio.Scanner, lineToWaitFor string) string {
	output := strings.Builder{}
	for stdoutScanner.Scan() {
		// Strip all terminal control sequences and skip empty lines
		line := removeAnsiEscapeSequences(stdoutScanner.Text())
		if line == "" {
			continue
		}

		// Record the output
		output.WriteString(line + "\n")

		// Once we've seen the line we wanted to wait for, we break.
		if strings.HasPrefix(line, lineToWaitFor) {
			break
		}
	}
	return output.String()
}

func removeAnsiEscapeSequences(input string) string {
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`\x1b\[[0-9;]*[JKmsu]`), // strip colors
		regexp.MustCompile(`\a`),                   // strip bell characters
		regexp.MustCompile(`\x1B]2;`),              // strip terminal title
		regexp.MustCompile(`sql-prompt`),           // strip 'sql-prompt'
		regexp.MustCompile(`> \x1b\[2D`),           // strip cursor back
		regexp.MustCompile(`\x1b\[1A.*`),           // strip cursor up
	}

	for _, regex := range regexes {
		input = regex.ReplaceAllString(input, "")
	}

	return strings.TrimSpace(input)
}

func (s *CLITestSuite) TestFlinkStatmentExceptionList() {
	tests := []CLITest{
		{args: "flink statement exception list my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/exception/list.golden"},
		{args: "flink statement exception list my-statement --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/exception/list-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
