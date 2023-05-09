package test

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestConnect() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect --help", fixture: "connect/help.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml -o json", fixture: "connect/cluster/create-json.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml -o yaml", fixture: "connect/cluster/create-yaml.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml", fixture: "connect/cluster/create.golden"},
		{args: "connect cluster delete lcc-123 --cluster lkc-123 --force", fixture: "connect/cluster/delete.golden"},
		{args: "connect cluster delete lcc-123 --cluster lkc-123", input: "az-connector\n", fixture: "connect/cluster/delete-prompt.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/cluster/describe-json.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/cluster/describe-yaml.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123", fixture: "connect/cluster/describe.golden"},
		{args: "connect cluster list --cluster lkc-123 -o json", fixture: "connect/cluster/list-json.golden"},
		{args: "connect cluster list --cluster lkc-123 -o yaml", fixture: "connect/cluster/list-yaml.golden"},
		{args: "connect cluster list --cluster lkc-123", fixture: "connect/cluster/list.golden"},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml", fixture: "connect/cluster/update.golden"},
		{args: "connect event describe", fixture: "connect/event-describe.golden"},

		// Tests based on new config
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json -o json", fixture: "connect/cluster/create-new-config-json.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json -o yaml", fixture: "connect/cluster/create-yaml.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-malformed-new.json", fixture: "connect/cluster/create-malformed-new.golden", exitCode: 1},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-malformed-old.json", fixture: "connect/cluster/create-malformed-old.golden", exitCode: 1},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json", fixture: "connect/cluster/update.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestConnectClusterPause() {
	tests := []CLITest{
		{args: "connect cluster pause --help", fixture: "connect/cluster/pause-help.golden"},
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
		{args: "connect cluster resume --help", fixture: "connect/cluster/resume-help.golden"},
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
		{args: "connect plugin --help", fixture: "connect/plugin/help.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o json", fixture: "connect/plugin/describe-json.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o yaml", fixture: "connect/plugin/describe-yaml.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123", fixture: "connect/plugin/describe.golden"},
		{args: "connect plugin list --cluster lkc-123 -o json", fixture: "connect/plugin/list-json.golden"},
		{args: "connect plugin list --cluster lkc-123 -o yaml", fixture: "connect/plugin/list-yaml.golden"},
		{args: "connect plugin list --cluster lkc-123", fixture: "connect/plugin/list.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestConnectPluginInstall() {
	s.setupConfluentPlatform()
	defer s.teardownConfluentPlatform()

	confluentHomeEmpty := "test/fixtures/input/connect/confluent-empty"
	confluentHome733 := "test/fixtures/input/connect/confluent-7.3.3"

	tests := []CLITest{
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run --force", env: []string{"CONFLUENT_HOME=" + confluentHomeEmpty}, fixture: "connect/plugin/install-dry-run.golden"},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip", env: []string{"CONFLUENT_HOME=test"}, fixture: "connect/plugin/install-platform-not-found.golden", exitCode: 1},
		{args: "connect plugin install bad-id-format", fixture: "connect/plugin/install-plugin-not-found.golden", exitCode: 1},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %s/share/confluent-hub-components --dry-run --force", confluentHomeEmpty), env: []string{"CONFLUENT_HOME=" + confluentHomeEmpty}},
		{args: "connect plugin install test/fixtures/input/connect/test-plugin.zip --dry-run --force", env: []string{"CONFLUENT_HOME=" + confluentHome733}, fixture: "connect/plugin/install-worker-update-dry-run.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %[1]s/share/confluent-hub-components --worker-configs %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run --force", confluentHome733), fixture: "connect/plugin/install-all-flags-dry-run.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %[1]s/share/confluent-hub-components --worker-configs %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run", confluentHome733), input: "y\n", fixture: "connect/plugin/install-all-flags-dry-run-accept.golden"},
		{args: fmt.Sprintf("connect plugin install test/fixtures/input/connect/test-plugin.zip --plugin-directory %[1]s/share/confluent-hub-components --worker-configs %[1]s/etc/kafka/connect-distributed.properties,%[1]s/etc/kafka/connect-standalone.properties --dry-run", confluentHome733), input: "n\n", fixture: "connect/plugin/install-all-flags-dry-run-refuse.golden", exitCode: 1},
	}

	for _, tt := range tests {
		tt.login = "platform"
		s.runIntegrationTest(tt)
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

func (s *CLITestSuite) setupConfluentPlatform() {
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

	err = os.MkdirAll("test/fixtures/input/connect/confluent-7.3.3/share/confluent-hub-components", 0755)
	req.NoError(err)
	err = os.MkdirAll("test/fixtures/input/connect/confluent-7.3.3/share/java/confluent-common", 0755)
	req.NoError(err)
	err = os.MkdirAll("test/fixtures/input/connect/confluent-7.3.3/etc/kafka", 0755)
	req.NoError(err)
	err = os.MkdirAll("test/fixtures/input/connect/confluent-empty/share/confluent-hub-components", 0755)
	req.NoError(err)
	err = os.MkdirAll("test/fixtures/input/connect/confluent-empty/share/java/confluent-common", 0755)
	req.NoError(err)

	connectDistributedFile, err := os.OpenFile("test/fixtures/input/connect/confluent-7.3.3/etc/kafka/connect-distributed.properties", os.O_CREATE|os.O_RDWR, 0644)
	req.NoError(err)
	defer connectDistributedFile.Close()
	_, err = connectDistributedFile.WriteString("plugin.path = /usr/share/java")
	req.NoError(err)

	connectStandaloneFile, err := os.OpenFile("test/fixtures/input/connect/confluent-7.3.3/etc/kafka/connect-standalone.properties", os.O_CREATE|os.O_RDWR, 0644)
	req.NoError(err)
	defer connectStandaloneFile.Close()
	_, err = connectStandaloneFile.WriteString("plugin.path = /usr/share/java")
	req.NoError(err)
}

func (s *CLITestSuite) teardownConfluentPlatform() {
	req := require.New(s.T())

	err := os.Remove("test/fixtures/input/connect/test-plugin.zip")
	req.NoError(err)

	err = os.RemoveAll("test/fixtures/input/connect/confluent-7.3.3")
	req.NoError(err)

	err = os.RemoveAll("test/fixtures/input/connect/confluent-empty")
	req.NoError(err)
}
