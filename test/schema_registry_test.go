package test

import (
	"fmt"

	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	schemaPath   = getInputFixturePath("schema-registry", "schema-example.json")
	metadataPath = getInputFixturePath("schema-registry", "schema-metadata.json")
	rulesetPath  = getInputFixturePath("schema-registry", "schema-ruleset.json")
)

func (s *CLITestSuite) TestSchemaRegistryCluster() {
	tests := []CLITest{
		{args: "schema-registry cluster enable --cloud gcp --geo us --package advanced -o json", fixture: "schema-registry/cluster/enable-json.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us --package essentials -o yaml", fixture: "schema-registry/cluster/enable-yaml.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us --package advanced", fixture: "schema-registry/cluster/enable.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo somethingwrong --package advanced", fixture: "schema-registry/cluster/enable-invalid-geo.golden", exitCode: 1},
		{args: "schema-registry cluster enable --cloud aws --geo us --package invalid-package", fixture: "schema-registry/cluster/enable-invalid-package.golden", exitCode: 1},
		{args: "schema-registry cluster enable --geo us --package essentials", fixture: "schema-registry/cluster/enable-missing-flag.golden", exitCode: 1},
		{args: fmt.Sprintf("schema-registry cluster delete --environment %s", testserver.SRApiEnvId), input: "y\n", fixture: "schema-registry/cluster/delete.golden"},
		{args: fmt.Sprintf("schema-registry cluster delete --environment %s", testserver.SRApiEnvId), input: "n\n", fixture: "schema-registry/cluster/delete-terminated.golden"},
		{args: fmt.Sprintf("schema-registry cluster delete --environment %s", testserver.SRApiEnvId), input: "invalid_confirmation\n", fixture: "schema-registry/cluster/delete-invalid-confirmation.golden", exitCode: 1},
		{args: "schema-registry cluster upgrade", fixture: "schema-registry/cluster/upgrade-missing-flag.golden", exitCode: 1},
		{args: "schema-registry cluster upgrade --package invalid-package", fixture: "schema-registry/cluster/upgrade-invalid-package.golden", exitCode: 1},
		{args: fmt.Sprintf("schema-registry cluster upgrade --package essentials --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/cluster/upgrade-current-package.golden"},
		{args: fmt.Sprintf("schema-registry cluster upgrade --package advanced --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/cluster/upgrade.golden"},
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
		{args: fmt.Sprintf("schema-registry compatibility validate --subject payments --version 1 --schema %s --environment %s", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/compatibility/validate.golden"},
		{args: fmt.Sprintf("schema-registry compatibility validate --subject payments --version 1 --schema %s --environment %s -o json", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/compatibility/validate-json.golden"},
		{args: fmt.Sprintf("schema-registry compatibility validate --subject payments --version 1 --schema %s --environment %s -o yaml", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/compatibility/validate-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryConfigDescribe() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry config describe --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/config/describe-global.golden"},
		{args: fmt.Sprintf("schema-registry config describe --api-key key --api-secret secret --environment %s -o json", testserver.SRApiEnvId), fixture: "schema-registry/config/describe-global-json.golden"},
		{args: fmt.Sprintf("schema-registry config describe --api-key key --api-secret secret --environment %s -o yaml", testserver.SRApiEnvId), fixture: "schema-registry/config/describe-global-yaml.golden"},
		{args: fmt.Sprintf("schema-registry config describe --subject payments --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/config/describe-subject.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryExporter() {
	exporterConfigPath := getInputFixturePath("schema-registry", "schema-exporter-config.txt")

	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry exporter list --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/list.golden"},
		{args: fmt.Sprintf(`schema-registry exporter create myexporter --subjects foo,bar --context-type AUTO --subject-format my-\\${subject} --config-file %s --api-key key --api-secret secret --environment %s`, exporterConfigPath, testserver.SRApiEnvId), fixture: "schema-registry/exporter/create.golden"},
		{args: fmt.Sprintf("schema-registry exporter describe myexporter --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/describe.golden"},
		{args: fmt.Sprintf(`schema-registry exporter update myexporter --subjects foo,bar,test --subject-format my-\\${subject} --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId), fixture: "schema-registry/exporter/update.golden"},
		{args: fmt.Sprintf("schema-registry exporter delete myexporter --api-key key --api-secret secret --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/exporter/delete.golden"},
		{args: fmt.Sprintf("schema-registry exporter delete myexporter --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), input: "myexporter\n", fixture: "schema-registry/exporter/delete-prompt.golden"},
		{args: fmt.Sprintf("schema-registry exporter get-status myexporter --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/get-status.golden"},
		{args: fmt.Sprintf("schema-registry exporter get-config myexporter --output json --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/get-config-json.golden"},
		{args: fmt.Sprintf("schema-registry exporter get-config myexporter --output yaml --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/get-config-yaml.golden"},
		{args: fmt.Sprintf("schema-registry exporter pause myexporter --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/pause.golden"},
		{args: fmt.Sprintf("schema-registry exporter resume myexporter --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/resume.golden"},
		{args: fmt.Sprintf("schema-registry exporter reset myexporter --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/exporter/reset.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistrySchema() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry schema create --subject payments --schema %s --api-key key --api-secret secret --environment %s", schemaPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/create.golden"},
		{args: fmt.Sprintf("schema-registry schema create --subject payments --schema %s --metadata %s --ruleset %s --api-key key --api-secret secret --environment %s", schemaPath, metadataPath, rulesetPath, testserver.SRApiEnvId), fixture: "schema-registry/schema/create.golden"},
		{args: fmt.Sprintf("schema-registry schema delete --subject payments --version latest --api-key key --api-secret secret --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/schema/delete.golden"},
		{args: fmt.Sprintf("schema-registry schema delete --subject payments --version all --api-key key --api-secret secret --environment %s --force", testserver.SRApiEnvId), fixture: "schema-registry/schema/delete-all.golden"},
		{args: "schema-registry schema describe --subject payments", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references --subject payments", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --version 1", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references --version 1", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --subject payments --version 1 123", exitCode: 1, fixture: "schema-registry/schema/describe-both-id-and-subject.golden"},
		{args: "schema-registry schema describe --show-references --subject payments --version 1 123", exitCode: 1, fixture: "schema-registry/schema/describe-both-id-and-subject.golden"},
		{args: fmt.Sprintf("schema-registry schema describe --subject payments --version 2 --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe.golden"},
		{args: fmt.Sprintf("schema-registry schema describe 10 --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe.golden"},
		{args: fmt.Sprintf("schema-registry schema describe 1001 --show-references --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe-refs-id.golden"},
		{args: fmt.Sprintf("schema-registry schema describe 1005 --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe-with-ruleset.golden"},
		{args: fmt.Sprintf("schema-registry schema describe --subject lvl0 --version 1 --show-references --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/describe-refs-subject.golden"},
		{args: fmt.Sprintf("schema-registry schema list --subject-prefix mysubject-1 --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/list-schemas-subject.golden"},
		{args: fmt.Sprintf("schema-registry schema list --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/schema/list-schemas-default.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistrySubject() {
	tests := []CLITest{
		{args: fmt.Sprintf("schema-registry subject list --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/list.golden"},
		{args: fmt.Sprintf("schema-registry subject describe testSubject --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/describe.golden"},
		{args: fmt.Sprintf("schema-registry subject update testSubject --compatibility BACKWARD --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/update-compatibility.golden"},
		{args: fmt.Sprintf("schema-registry subject update testSubject --compatibility BACKWARD --compatibility-group application.version --metadata-defaults %s --ruleset-defaults %s --api-key key --api-secret secret --environment %s", metadataPath, rulesetPath, testserver.SRApiEnvId), fixture: "schema-registry/subject/update-compatibility.golden"},
		{args: fmt.Sprintf("schema-registry subject update testSubject --mode readonly --api-key key --api-secret secret --environment %s", testserver.SRApiEnvId), fixture: "schema-registry/subject/update-mode.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestSchemaRegistryRegionList() {
	tests := []CLITest{
		{args: "schema-registry region list", fixture: "schema-registry/region/list-all.golden"},
		{args: "schema-registry region list -o json", fixture: "schema-registry/region/list-all-json.golden"},
		{args: "schema-registry region list --cloud aws", fixture: "schema-registry/region/list-filter-cloud.golden"},
		{args: "schema-registry region list --package advanced", fixture: "schema-registry/region/list-filter-package.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
