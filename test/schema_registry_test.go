package test

import (
	"fmt"

	testserver "github.com/confluentinc/cli/test/test-server"
)

func (s *CLITestSuite) TestSchemaRegistry() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	schemaPath := GetInputFixturePath(s.T(), "schema-registry", "schema-example.json")
	exporterConfigPath := GetInputFixturePath(s.T(), "schema-registry", "schema-exporter-config.txt")
	metadataPath := GetInputFixturePath(s.T(), "schema-registry", "schema-metadata.json")
	rulesetPath := GetInputFixturePath(s.T(), "schema-registry", "schema-ruleset.json")

	tests := []CLITest{
		{args: "schema-registry --help", fixture: "schema-registry/help.golden"},
		{args: "schema-registry cluster --help", fixture: "schema-registry/cluster/help.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us --package advanced -o json", fixture: "schema-registry/cluster/enable-json.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us --package essentials -o yaml", fixture: "schema-registry/cluster/enable-yaml.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us --package advanced", fixture: "schema-registry/cluster/enable.golden"},
		{
			args:     "schema-registry cluster enable --cloud gcp --geo somethingwrong --package advanced",
			fixture:  "schema-registry/cluster/enable-invalid-geo.golden",
			exitCode: 1,
		},
		{
			args:     "schema-registry cluster enable --cloud aws --geo us --package invalid-package",
			fixture:  "schema-registry/cluster/enable-invalid-package.golden",
			exitCode: 1,
		},
		{
			args:     "schema-registry cluster enable --geo us --package essentials",
			fixture:  "schema-registry/cluster/enable-missing-flag.golden",
			exitCode: 1,
		},
		{
			args:    fmt.Sprintf(`schema-registry cluster delete --environment %s`, testserver.SRApiEnvId),
			input:   "y\n",
			fixture: "schema-registry/cluster/delete.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry cluster delete --environment %s`, testserver.SRApiEnvId),
			input:   "n\n",
			fixture: "schema-registry/cluster/delete-terminated.golden",
		},
		{
			args:     fmt.Sprintf(`schema-registry cluster delete --environment %s`, testserver.SRApiEnvId),
			input:    "invalid_confirmation\n",
			fixture:  "schema-registry/cluster/delete-invalid-confirmation.golden",
			exitCode: 1,
		},
		{
			args:     "schema-registry cluster upgrade",
			fixture:  "schema-registry/cluster/upgrade-missing-flag.golden",
			exitCode: 1,
		},
		{
			args:     "schema-registry cluster upgrade --package invalid-package",
			fixture:  "schema-registry/cluster/upgrade-invalid-package.golden",
			exitCode: 1,
		},
		{
			args:    fmt.Sprintf(`schema-registry cluster upgrade --package essentials --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/cluster/upgrade-current-package.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry cluster upgrade --package advanced --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/cluster/upgrade.golden",
		},
		{args: "schema-registry schema --help", fixture: "schema-registry/schema/help.golden"},
		{args: "schema-registry subject --help", fixture: "schema-registry/subject/help.golden"},
		{args: "schema-registry exporter --help", fixture: "schema-registry/exporter/help.golden"},
		{args: "schema-registry exporter create --help", fixture: "schema-registry/exporter/create-help.golden"},
		{args: "schema-registry exporter delete --help", fixture: "schema-registry/exporter/delete-help.golden"},
		{args: "schema-registry exporter describe --help", fixture: "schema-registry/exporter/describe-help.golden"},
		{args: "schema-registry exporter get-config --help", fixture: "schema-registry/exporter/get-config-help.golden"},
		{args: "schema-registry exporter get-status --help", fixture: "schema-registry/exporter/get-status-help.golden"},
		{args: "schema-registry exporter list --help", fixture: "schema-registry/exporter/list-help.golden"},
		{args: "schema-registry exporter pause --help", fixture: "schema-registry/exporter/pause-help.golden"},
		{args: "schema-registry exporter reset --help", fixture: "schema-registry/exporter/reset-help.golden"},
		{args: "schema-registry exporter resume --help", fixture: "schema-registry/exporter/resume-help.golden"},
		{args: "schema-registry exporter update --help", fixture: "schema-registry/exporter/update-help.golden"},
		{args: "schema-registry cluster update --help", fixture: "schema-registry/cluster/update-help.golden"},
		{args: "schema-registry subject update --help", fixture: "schema-registry/subject/update-help.golden"},

		{args: "schema-registry cluster describe", fixture: "schema-registry/cluster/describe.golden"},
		{
			args:     fmt.Sprintf(`schema-registry cluster update --environment %s`, testserver.SRApiEnvId),
			fixture:  "schema-registry/cluster/update-missing-flags.golden",
			exitCode: 1,
		},
		{
			args:    fmt.Sprintf(`schema-registry cluster update --compatibility BACKWARD --environment %s`, testserver.SRApiEnvId),
			input:   "key\nsecret\n",
			fixture: "schema-registry/cluster/update-compatibility.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry cluster update --mode READWRITE --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/cluster/update-mode.golden",
		},
		{
			name:    "schema-registry schema create",
			args:    fmt.Sprintf(`schema-registry schema create --subject payments --schema %s --api-key key --api-secret secret --environment %s`, schemaPath, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/create.golden",
		},
		{
			name:    "schema-registry schema create with metadata and ruleset",
			args:    fmt.Sprintf(`schema-registry schema create --subject payments --schema %s --metadata %s --ruleset %s --api-key key --api-secret secret --environment %s`, schemaPath, metadataPath, rulesetPath, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/create.golden",
		},
		{
			name:    "schema-registry compatibility validate",
			args:    fmt.Sprintf(`schema-registry compatibility validate --subject payments --version 1 --schema %s --api-key key --api-secret secret --environment %s`, schemaPath, testserver.SRApiEnvId),
			fixture: "schema-registry/compatibility/validate.golden",
		},
		{
			name:    "schema-registry compatibility validate json",
			args:    fmt.Sprintf(`schema-registry compatibility validate --subject payments --version 1 --schema %s --api-key key --api-secret secret --environment %s -o json`, schemaPath, testserver.SRApiEnvId),
			fixture: "schema-registry/compatibility/validate-json.golden",
		},
		{
			name:    "schema-registry compatibility validate yaml",
			args:    fmt.Sprintf(`schema-registry compatibility validate --subject payments --version 1 --schema %s --api-key key --api-secret secret --environment %s -o yaml`, schemaPath, testserver.SRApiEnvId),
			fixture: "schema-registry/compatibility/validate-yaml.golden",
		},
		{
			name:    "schema-registry config describe global",
			args:    fmt.Sprintf(`schema-registry config describe --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/config/describe-global.golden",
		},
		{
			name:    "schema-registry config describe global json",
			args:    fmt.Sprintf(`schema-registry config describe --api-key key --api-secret secret --environment %s -o json`, testserver.SRApiEnvId),
			fixture: "schema-registry/config/describe-global-json.golden",
		},
		{
			name:    "schema-registry config describe global yaml",
			args:    fmt.Sprintf(`schema-registry config describe --api-key key --api-secret secret --environment %s -o yaml`, testserver.SRApiEnvId),
			fixture: "schema-registry/config/describe-global-yaml.golden",
		},
		{
			name:    "schema-registry config describe --subject payments",
			args:    fmt.Sprintf(`schema-registry config describe --subject payments --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/config/describe-subject.golden",
		},
		{
			name:    "schema-registry schema delete latest",
			args:    fmt.Sprintf(`schema-registry schema delete --subject payments --version latest --api-key key --api-secret secret --environment %s --force`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/delete.golden",
		},
		{
			name:    "schema-registry schema delete all",
			args:    fmt.Sprintf(`schema-registry schema delete --subject payments --version all --api-key key --api-secret secret --environment %s --force`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/delete-all.golden",
		},
		{args: "schema-registry schema describe --subject payments", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references --subject payments", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --version 1", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references --version 1", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --show-references", exitCode: 1, fixture: "schema-registry/schema/describe-either-id-or-subject.golden"},
		{args: "schema-registry schema describe --subject payments --version 1 123", exitCode: 1, fixture: "schema-registry/schema/describe-both-id-and-subject.golden"},
		{args: "schema-registry schema describe --show-references --subject payments --version 1 123", exitCode: 1, fixture: "schema-registry/schema/describe-both-id-and-subject.golden"},
		{
			name:    "schema-registry schema describe --subject payments --version 2",
			args:    fmt.Sprintf(`schema-registry schema describe --subject payments --version 2 --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/describe.golden",
		},
		{
			name:    "schema-registry schema describe by id",
			args:    fmt.Sprintf(`schema-registry schema describe 10 --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/describe.golden",
		},
		{
			name:    "schema-registry schema describe 1001 --show-references",
			args:    fmt.Sprintf(`schema-registry schema describe 1001 --show-references --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/describe-refs-id.golden",
		},
		{
			name:    "schema-registry schema describe 1005",
			args:    fmt.Sprintf(`schema-registry schema describe 1005 --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/describe-with-ruleset.golden",
		},
		{
			name:    "schema-registry schema describe --subject lvl0 --version 1 --show-references",
			args:    fmt.Sprintf(`schema-registry schema describe --subject lvl0 --version 1 --show-references --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/describe-refs-subject.golden",
		},
		{
			name:    "schema-registry schema list --subject-prefix mysubject-1",
			args:    fmt.Sprintf(`schema-registry schema list --subject-prefix mysubject-1 --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/list-schemas-subject.golden",
		},
		{
			name:    "schema-registry schema list",
			args:    fmt.Sprintf(`schema-registry schema list --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/schema/list-schemas-default.golden",
		},
		{
			name:    "schema-registry subject list",
			args:    fmt.Sprintf(`schema-registry subject list --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/subject/list.golden",
		},
		{
			name:    "schema-registry subject describe testSubject",
			args:    fmt.Sprintf(`schema-registry subject describe testSubject --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/subject/describe.golden",
		},
		{
			name:    "schema-registry subject update compatibility",
			args:    fmt.Sprintf(`schema-registry subject update testSubject --compatibility BACKWARD --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/subject/update-compatibility.golden",
		},
		{
			name:    "schema-registry subject update compatibility with metadata and ruleset",
			args:    fmt.Sprintf(`schema-registry subject update testSubject --compatibility BACKWARD --compatibility-group application.version --metadata-defaults %s --ruleset-defaults %s --api-key key --api-secret secret --environment %s`, metadataPath, rulesetPath, testserver.SRApiEnvId),
			fixture: "schema-registry/subject/update-compatibility.golden",
		},
		{
			name:    "schema-registry subject update mode",
			args:    fmt.Sprintf(`schema-registry subject update testSubject --mode READONLY --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/subject/update-mode.golden",
		},

		{
			args:    fmt.Sprintf(`schema-registry exporter list --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/list.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter create myexporter --subjects foo,bar --context-type AUTO --subject-format my-\\${subject} --config-file %s --api-key key --api-secret secret --environment %s`, exporterConfigPath, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/create.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter describe myexporter --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/describe.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter update myexporter --subjects foo,bar,test --subject-format my-\\${subject} --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/update.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter get-status myexporter --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/get-status.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter get-config myexporter --output json --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/get-config-json.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter get-config myexporter --output yaml --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/get-config-yaml.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter pause myexporter --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/pause.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter resume myexporter --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/resume.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter reset myexporter --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/reset.golden",
		},
		{
			args:    "schema-registry region --help",
			fixture: "schema-registry/region/help.golden",
		},
		{
			args:    "schema-registry region list",
			fixture: "schema-registry/region/list-all.golden",
		},
		{
			args:    "schema-registry region list -o json",
			fixture: "schema-registry/region/list-all-json.golden",
		},
		{
			args:    "schema-registry region list --cloud aws",
			fixture: "schema-registry/region/list-filter-cloud.golden",
		},
		{
			args:    "schema-registry region list --package advanced",
			fixture: "schema-registry/region/list-filter-package.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestSchemaRegistryExporterDelete() {
	tests := []CLITest{
		{
			args:    fmt.Sprintf(`schema-registry exporter delete myexporter --api-key key --api-secret secret --environment %s --force`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/delete.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter delete myexporter myexporter2 --api-key key --api-secret secret --environment %s --force`, testserver.SRApiEnvId),
			fixture: "schema-registry/exporter/delete-multiple-success.golden",
		},
		{
			args:    fmt.Sprintf(`schema-registry exporter delete myexporter --api-key key --api-secret secret --environment %s`, testserver.SRApiEnvId),
			input:   "myexporter\n",
			fixture: "schema-registry/exporter/delete-prompt.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
