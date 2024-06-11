package test

import (
	"fmt"

	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

var (
	schemaPath   = getInputFixturePath("schema-registry", "schema-example.json")
	metadataPath = getInputFixturePath("schema-registry", "schema-metadata.json")
	rulesetPath  = getInputFixturePath("schema-registry", "schema-ruleset.json")
)

func (s *CLITestSuite) TestSchemaRegistryCluster() {
	tests := []CLITest{
		{args: "schema-registry cluster describe", fixture: "schema-registry/cluster/describe.golden"},
		{args: fmt.Sprintf("schema-registry cluster update --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/cluster/update-missing-flags.golden", exitCode: 1},
		{args: fmt.Sprintf("schema-registry cluster update --compatibility backward --environment %s", testserver.SRApiEnvId), input: "key\nsecret\n", fixture: "schema-registry/cluster/update-compatibility.golden"},
		{args: fmt.Sprintf("schema-registry cluster update --mode readwrite --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/cluster/update-mode.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryCompatibilityValidate() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry schema compatibility validate --subject payments --version 1 --schema %s --environment %s", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/compatibility/validate.golden"},
		{args: fmt.Sprintf("schema-registry schema compatibility validate --subject payments --version 1 --schema %s --environment %s -o json", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/compatibility/validate-json.golden"},
		{args: fmt.Sprintf("schema-registry schema compatibility validate --subject payments --version 1 --schema %s --environment %s -o yaml", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/compatibility/validate-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryConfigDescribe() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry configuration describe --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/configuration/describe-global.golden"},
		{args: fmt.Sprintf("schema-registry configuration describe --environment %s -o json", testserver.SRApiEnvId), fixture: "schema-registry/configuration/describe-global-json.golden"},
		{args: fmt.Sprintf("schema-registry configuration describe --environment %s -o yaml", testserver.SRApiEnvId), fixture: "schema-registry/configuration/describe-global-yaml.golden"},
		{args: fmt.Sprintf("schema-registry configuration describe --subject payments --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/configuration/describe-subject.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryConfigDelete() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry configuration delete --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/configuration/delete.golden"},
		{args: fmt.Sprintf("schema-registry configuration delete --subject payments --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/configuration/delete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryExporter() {
	exporterConfigPath := getInputFixturePath("schema-registry", "schema-exporter-config.txt")

	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry exporter list --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/list.golden"},
		{args: fmt.Sprintf(`schema-registry exporter create myexporter --subjects foo,bar --context-type AUTO --subject-format my-\\${subject} --config %s --environment %s`, exporterConfigPath, testserver.SRApiEnvId), fixture: "schema-registry/exporter/create.golden"},
		{args: fmt.Sprintf("schema-registry exporter describe myexporter --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/describe.golden"},
		{args: fmt.Sprintf(`schema-registry exporter update myexporter --subjects foo,bar,test --subject-format my-\\${subject} --environment %s`, testserver.SRApiEnvId), fixture: "schema-registry/exporter/update.golden"},
		{args: fmt.Sprintf("schema-registry exporter delete myexporter --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/exporter/delete.golden"},
		{args: fmt.Sprintf(`schema-registry exporter delete myexporter myexporter2 --environment %s`, testserver.SRApiEnvId), input: "n\n", fixture: "schema-registry/exporter/delete-multiple-refuse.golden"},
		{args: fmt.Sprintf(`schema-registry exporter delete myexporter myexporter2 --environment %s --force`, testserver.SRApiEnvId), fixture: "schema-registry/exporter/delete-multiple-success.golden"},
		{args: fmt.Sprintf("schema-registry exporter delete myexporter --environment %s", testserver.SRApiEnvId), input: "y\n", fixture: "schema-registry/exporter/delete-prompt.golden"},
		{args: fmt.Sprintf("schema-registry exporter status describe myexporter --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/status/describe.golden"},
		{args: fmt.Sprintf("schema-registry exporter configuration describe myexporter --output json --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/configuration/describe-json.golden"},
		{args: fmt.Sprintf("schema-registry exporter configuration describe myexporter --output yaml --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/configuration/describe-yaml.golden"},
		{args: fmt.Sprintf("schema-registry exporter pause myexporter --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/pause.golden"},
		{args: fmt.Sprintf("schema-registry exporter resume myexporter --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/resume.golden"},
		{args: fmt.Sprintf("schema-registry exporter reset myexporter --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/reset.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistrySchema() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry schema create --subject payments --schema %s --environment %s", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/create.golden"},
		{args: fmt.Sprintf("schema-registry schema create --subject payments --schema %s --metadata %s --ruleset %s --environment %s", schemaPath, metadataPath, rulesetPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/create.golden"},
		{args: fmt.Sprintf("schema-registry schema delete --subject payments --version latest --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/schema/delete.golden"},
		{args: fmt.Sprintf("schema-registry schema delete --subject payments --version all --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/schema/delete-all.golden"},
		{args: "schema-registry schema describe --subject payments", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references --subject payments", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --version 1", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references --version 1", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --subject payments --version 1 123", exitCode: 1, fixture: "schema-registry/schema/describe-both-id-and-subject.golden"},
		{args: "schema-registry schema describe --show-references --subject payments --version 1 123", exitCode: 1, fixture: "schema-registry/schema/describe-both-id-and-subject.golden"},
		{args: fmt.Sprintf("schema-registry schema describe --subject payments --version 2 --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe.golden"},
		{args: fmt.Sprintf("schema-registry schema describe 10 --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe.golden"},
		{args: fmt.Sprintf("schema-registry schema describe 1001 --show-references --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe-refs-id.golden"},
		{args: fmt.Sprintf("schema-registry schema describe 1005 --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe-with-ruleset.golden"},
		{args: fmt.Sprintf("schema-registry schema describe --subject lvl0 --version 1 --show-references --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe-refs-subject.golden"},
		{args: fmt.Sprintf("schema-registry schema list --subject-prefix mysubject-1 --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/list-schemas-subject.golden"},
		{args: fmt.Sprintf("schema-registry schema list --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/list-schemas-default.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistrySubject() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry subject list --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/list.golden"},
		{args: fmt.Sprintf("schema-registry subject describe testSubject --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/describe.golden"},
		{args: fmt.Sprintf("schema-registry subject update testSubject --compatibility backward --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/update-compatibility.golden"},
		{args: fmt.Sprintf("schema-registry subject update testSubject --compatibility backward --compatibility-group application.version --metadata-defaults %s --ruleset-defaults %s --environment %s", metadataPath, rulesetPath, testserver.SRApiEnvId), fixture: "schema-registry/subject/update-compatibility.golden"},
		{args: fmt.Sprintf("schema-registry subject update testSubject --mode readonly --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/update-mode.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryKek() {
	tests := []CLITest{
		{args: "schema-registry kek create --name kek-name --kms-type AWS_KMS --kms-key arn:aws:kms:us-west-2:9979:key/abcd --kms-properties KeyState=Enabled --doc description", fixture: "schema-registry/kek/create.golden"},
		{args: "schema-registry kek list -o json", fixture: "schema-registry/kek/list-all-json.golden"},
		{args: "schema-registry kek describe kek-name", fixture: "schema-registry/kek/describe.golden"},
		{args: "schema-registry kek update kek-name --doc new-description", fixture: "schema-registry/kek/update.golden"},
		{args: "schema-registry kek delete kek-name --force", fixture: "schema-registry/region/delete.golden"},
		{args: "schema-registry kek undelete kek-name --force", fixture: "schema-registry/region/undelete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryDek() {
	tests := []CLITest{
		{args: "schema-registry dek create --kek-name kek-name --subject payments --algorithm AES256_GCM --version 1 --encrypted-key-material encrypted-key-material", fixture: "schema-registry/dek/create.golden"},
		{args: "schema-registry dek subject list --kek-name kek-name", fixture: "schema-registry/dek/list-subject.golden"},
		{args: "schema-registry dek version list --kek-name kek-name --subject payments", fixture: "schema-registry/dek/list-version.golden"},
		{args: "schema-registry dek describe --kek-name kek-name --subject payments", fixture: "schema-registry/dek/describe.golden"},
		{args: "schema-registry dek delete --kek-name kek-name --subject payments --force", fixture: "schema-registry/dek/delete.golden"},
		{args: "schema-registry dek undelete --kek-name kek-name --subject payments --version 2", fixture: "schema-registry/dek/undelete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
