package test

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestConnect() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml -o json", fixture: "connect/cluster/create-json.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml -o yaml", fixture: "connect/cluster/create-yaml.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml", fixture: "connect/cluster/create.golden"},
		{args: "connect cluster delete lcc-123 --cluster lkc-123 --force", fixture: "connect/cluster/delete.golden"},
		{args: "connect cluster delete lcc-123 --cluster lkc-123", input: "az-connector\n", fixture: "connect/cluster/delete-prompt.golden"},
		{args: "connect cluster delete lcc-123 lcc-456 lcc-789 lcc-101112 --cluster lkc-123", input: "y\n", fixture: "connect/cluster/delete-multiple-fail.golden", exitCode: 1},
		{args: "connect cluster delete lcc-123 lcc-111 --cluster lkc-123", input: "n\n", fixture: "connect/cluster/delete-multiple-refuse.golden"},
		{args: "connect cluster delete lcc-123 lcc-111 --cluster lkc-123", input: "y\n", fixture: "connect/cluster/delete-multiple-success.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/cluster/describe-json.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/cluster/describe-yaml.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123", fixture: "connect/cluster/describe.golden"},
		{args: "connect cluster list --cluster lkc-123 -o json", fixture: "connect/cluster/list-json.golden"},
		{args: "connect cluster list --cluster lkc-123 -o yaml", fixture: "connect/cluster/list-yaml.golden"},
		{args: "connect cluster list --cluster lkc-123", fixture: "connect/cluster/list.golden"},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/update-config.yaml", fixture: "connect/cluster/update.golden"},
		{args: "connect event describe", fixture: "connect/event-describe.golden"},

		// Tests based on new config
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json -o json", fixture: "connect/cluster/create-new-config-json.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json -o yaml", fixture: "connect/cluster/create-yaml.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-malformed-new.json", fixture: "connect/cluster/create-malformed-new.golden", exitCode: 1},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-malformed-old.json", fixture: "connect/cluster/create-malformed-old.golden", exitCode: 1},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/update-config-new-format.json", fixture: "connect/cluster/update.golden"},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/update-config-malformed.json", fixture: "connect/cluster/update-error.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectClusterPause() {
	tests := []CLITest{
		{args: "connect cluster pause lcc-000000 --cluster lkc-123456", fixture: "connect/cluster/pause-unknown.golden", exitCode: 1},
		{args: "connect cluster pause lcc-123 --cluster lkc-123456", fixture: "connect/cluster/pause.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectClusterResume() {
	tests := []CLITest{
		{args: "connect cluster resume lcc-000000 --cluster lkc-123456", fixture: "connect/cluster/resume-unknown.golden", exitCode: 1},
		{args: "connect cluster resume lcc-123 --cluster lkc-123456", fixture: "connect/cluster/resume.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectPlugin() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o json", fixture: "connect/plugin/describe-json.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o yaml", fixture: "connect/plugin/describe-yaml.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123", fixture: "connect/plugin/describe.golden"},
		{args: "connect plugin list --cluster lkc-123 -o json", fixture: "connect/plugin/list-json.golden"},
		{args: "connect plugin list --cluster lkc-123 -o yaml", fixture: "connect/plugin/list-yaml.golden"},
		{args: "connect plugin list --cluster lkc-123", fixture: "connect/plugin/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectPluginInstall() {
	if runtime.GOOS == "windows" {
		return
	}

	s.zipManifest()
	defer s.deleteZip()

	confluentHome733 := "test/fixtures/input/connect/confluent-7.3.3"
	confluentHomeEmpty := "test/fixtures/input/connect/confluent-empty"
	confluentHomePriorInstall := "test/fixtures/input/connect/confluent-prior-install"

	//nolint:dupword
	tests := []CLITest{
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\ny\n", fixture: "connect/plugin/install/interactive.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\nn\ny\nn\n", fixture: "connect/plugin/install/interactive-select-workers.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHomeEmpty}, input: "y\ny\n", fixture: "connect/plugin/install/interactive-no-workers.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: fmt.Sprintf("n\n%s/share\ny\ny\n", confluentHomeEmpty), fixture: "connect/plugin/install/interactive-select-plugin-directory.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "n\n/directory-dne\ny\ny\n", fixture: "connect/plugin/install/interactive-select-plugin-directory-fail.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\nn\n", fixture: "connect/plugin/install/interactive-decline-license.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run --force", env: []string{"CONFLUENT_HOME=" + confluentHome733}, fixture: "connect/plugin/install/force.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHomePriorInstall}, input: "y\ny\ny\ny\n", fixture: "connect/plugin/install/interactive-uninstall-prior.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHomePriorInstall}, input: "y\nn\n", fixture: "connect/plugin/install/interactive-decline-uninstall-prior.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run --force", env: []string{"CONFLUENT_HOME=" + confluentHomePriorInstall}, fixture: "connect/plugin/install/uninstall-prior-force.golden"},

		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %s/share/confluent-hub-components --dry-run", confluentHome733), env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\n", fixture: "connect/plugin/install/plugin-directory-flag.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %s/share/confluent-hub-components --dry-run --force", confluentHome733), env: []string{"CONFLUENT_HOME=" + confluentHome733}, fixture: "connect/plugin/install/plugin-directory-flag-force.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --worker-configurations %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run", confluentHome733), env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\n", fixture: "connect/plugin/install/worker-configurations-flag.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --worker-configurations %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run --force", confluentHome733), env: []string{"CONFLUENT_HOME=" + confluentHome733}, fixture: "connect/plugin/install/worker-configurations-flag-force.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %[1]s/share/confluent-hub-components --worker-configurations %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run", confluentHome733), input: "y\n", fixture: "connect/plugin/install/both-flags.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %[1]s/share/confluent-hub-components --worker-configurations %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run --force", confluentHome733), fixture: "connect/plugin/install/both-flags-force.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --confluent-platform %s --dry-run", confluentHome733), input: "y\ny\ny\n", fixture: "connect/plugin/install/platform-flag.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --confluent-platform test/fixtures --dry-run", fixture: "connect/plugin/install/platform-flag-fail.golden", exitCode: 1},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --confluent-platform %[1]s --plugin-directory %[1]s/share/confluent-hub-components --worker-configurations %[1]s/etc/kafka/connect-distributed.properties", confluentHome733), fixture: "connect/plugin/install/all-file-flags.golden", exitCode: 1},

		{args: "connect plugin install bad-id-format", fixture: "connect/plugin/install/plugin-not-found.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip", env: []string{"CONFLUENT_HOME=test"}, fixture: "connect/plugin/install/platform-not-found.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory /directory-dne", fixture: "connect/plugin/install/plugin-directory-not-found.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --worker-configurations /dne.properties", fixture: "connect/plugin/install/worker-configurations-not-found.golden", exitCode: 1},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --worker-configurations /dne.properties,/dne2.properties", fixture: "connect/plugin/install/worker-configurations-multiple-not-found.golden", exitCode: 1},

		{args: "connect plugin install confluentinc/integration-test-plugin:latest --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\ny\n", fixture: "connect/plugin/install/remote.golden"},
		{args: "connect plugin install confluentinc/integration-test-plugin:0.0.5 --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\ny\n", fixture: "connect/plugin/install/remote-specific-version.golden"},
		{args: "connect plugin install confluentinc/dne-connector:latest --dry-run", env: []string{"CONFLUENT_HOME=" + confluentHome733}, fixture: "connect/plugin/install/remote-dne.golden", exitCode: 1},
		{args: "connect plugin install confluentinc/bad-md5:latest", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\n", fixture: "connect/plugin/install/remote-bad-md5.golden", exitCode: 1},
		{args: "connect plugin install confluentinc/bad-sha1:latest", env: []string{"CONFLUENT_HOME=" + confluentHome733}, input: "y\ny\n", fixture: "connect/plugin/install/remote-bad-sha1.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnect_Autocomplete() {
	tests := []CLITest{
		{args: `__complete connect cluster describe ""`, useKafka: "lkc-123", fixture: "connect/cluster/describe-autocomplete.golden"},
		{args: `__complete connect plugin describe ""`, useKafka: "lkc-123", fixture: "connect/plugin/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) zipManifest() {
	req := require.New(s.T())

	zipFile, err := os.Create("test/fixtures/input/connect/test-plugin.zip")
	req.NoError(err)
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	manifestFile, err := os.Open("test/fixtures/input/connect/manifest.json")
	req.NoError(err)
	defer manifestFile.Close()

	zipManifest, err := writer.Create("test-plugin/manifest.json")
	req.NoError(err)
	_, err = io.Copy(zipManifest, manifestFile)
	req.NoError(err)
}

func (s *CLITestSuite) deleteZip() {
	req := require.New(s.T())

	err := os.Remove("test/fixtures/input/connect/test-plugin.zip")
	req.NoError(err)
}

func (s *CLITestSuite) TestConnectCustomPlugin() {
	tests := []CLITest{
		{args: `connect custom-plugin create my-custom-plugin --plugin-file "test/fixtures/input/connect/confluentinc-kafka-connect-datagen-0.6.1.zip" --connector-type source --connector-class io.confluent.kafka.connect.datagen.DatagenConnector --cloud aws`, fixture: "connect/custom-plugin/create.golden"},
		{args: `connect custom-plugin create my-custom-plugin --plugin-file "test/fixtures/input/connect/confluentinc-kafka-connect-datagen-0.6.1.zip" --connector-type source --connector-class io.confluent.kafka.connect.datagen.DatagenConnector`, fixture: "connect/custom-plugin/create.golden"},
		{args: `connect custom-plugin create my-custom-plugin --plugin-file "test/fixtures/input/connect/confluentinc-kafka-connect-datagen-0.6.1.pdf" --connector-type source --connector-class io.confluent.kafka.connect.datagen.DatagenConnector --cloud aws`, fixture: "connect/custom-plugin/create-invalid-extension.golden", exitCode: 1},
		{args: "connect custom-plugin list", fixture: "connect/custom-plugin/list.golden"},
		{args: "connect custom-plugin list --cloud aws", fixture: "connect/custom-plugin/list.golden"},
		{args: "connect custom-plugin list -o json", fixture: "connect/custom-plugin/list-json.golden"},
		{args: "connect custom-plugin list -o yaml", fixture: "connect/custom-plugin/list-yaml.golden"},
		{args: "connect custom-plugin describe ccp-123456", fixture: "connect/custom-plugin/describe.golden"},
		{args: "connect custom-plugin describe ccp-789012", fixture: "connect/custom-plugin/describe-with-sensitive-properties.golden"},
		{args: "connect custom-plugin describe ccp-123456 -o json", fixture: "connect/custom-plugin/describe-json.golden"},
		{args: "connect custom-plugin describe ccp-123456 -o yaml", fixture: "connect/custom-plugin/describe-yaml.golden"},
		{args: "connect custom-plugin delete ccp-123456 --force", fixture: "connect/custom-plugin/delete.golden"},
		{args: "connect custom-plugin delete ccp-123456", input: "CliPluginTest1\n", fixture: "connect/custom-plugin/delete-prompt.golden"},
		{args: "connect custom-plugin update ccp-123456 --name CliPluginTestUpdate", fixture: "connect/custom-plugin/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectOffset() {
	tests := []CLITest{
		{args: "connect offset describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/offset/describe-offset-json.golden"},
		{args: "connect offset describe lcc-101112 --cluster lkc-123 -o json", fixture: "connect/offset/describe-offset-fail.golden", exitCode: 1},
		{args: "connect offset describe lcc-123 --cluster lkc-123", fixture: "connect/offset/describe-offset.golden"},
		{args: "connect offset describe lcc-123 --staleness-threshold 10 --timeout 10 --cluster lkc-123", fixture: "connect/offset/describe-offset.golden"},

		{args: "connect offset describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/offset/describe-offset-yaml.golden"},
		{args: "connect offset update lcc-123 --config-file test/fixtures/input/connect/offset.json --cluster lkc-123", fixture: "connect/offset/update-offset.golden"},
		{args: "connect offset update lcc-123 --config-file test/fixtures/input/connect/offset.json --cluster lkc-123 -o json", fixture: "connect/offset/update-offset-json.golden"},
		{args: "connect offset update lcc-123 --config-file test/fixtures/input/connect/offset.json --cluster lkc-123 -o yaml", fixture: "connect/offset/update-offset-yaml.golden"},

		{args: "connect offset status describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/offset/update-offset-status-describe-json.golden"},
		{args: "connect offset status describe lcc-123 --cluster lkc-123", fixture: "connect/offset/update-offset-status-describe.golden"},
		{args: "connect offset status describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/offset/update-offset-status-describe-yaml.golden"},

		{args: "connect offset delete lcc-111 --cluster lkc-123", fixture: "connect/offset/delete-offset.golden"},
		{args: "connect offset delete lcc-111 --cluster lkc-123 -o json", fixture: "connect/offset/delete-offset-json.golden"},
		{args: "connect offset delete lcc-111 --cluster lkc-123 -o yaml", fixture: "connect/offset/delete-offset-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
