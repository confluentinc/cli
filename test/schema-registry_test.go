package test

import (
	"strings"

	"github.com/confluentinc/bincover"

	testserver "github.com/confluentinc/cli/test/test-server"
)

func (s *CLITestSuite) TestSchemaRegistry() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	schemaPath := GetInputFixturePath(s.T(), "schema", "schema-example.json")
	exporterConfigPath := GetInputFixturePath(s.T(), "schema", "schema-exporter-config.txt")

	tests := []CLITest{
		{args: "schema-registry --help", fixture: "schema-registry/schema-registry-help.golden"},
		{args: "schema-registry cluster --help", fixture: "schema-registry/schema-registry-cluster-help.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o json", fixture: "schema-registry/schema-registry-enable-json.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o yaml", fixture: "schema-registry/schema-registry-enable-yaml.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us", fixture: "schema-registry/schema-registry-enable.golden"},
		{args: "schema-registry schema --help", fixture: "schema-registry/schema-registry-schema-help.golden"},
		{args: "schema-registry subject --help", fixture: "schema-registry/schema-registry-subject-help.golden"},
		{args: "schema-registry exporter --help", fixture: "schema-registry/schema-registry-exporter-help.golden"},

		{args: "schema-registry cluster describe", fixture: "schema-registry/schema-registry-describe.golden"},
		{args: "schema-registry cluster update --environment=" + testserver.SRApiEnvId, fixture: "schema-registry/schema-registry-update-missing-flags.golden", wantErrCode: 1},
		{args: "schema-registry cluster update --compatibility BACKWARD --environment=" + testserver.SRApiEnvId, preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("key\nsecret\n"))}, fixture: "schema-registry/schema-registry-update-compatibility.golden"},
		{args: "schema-registry cluster update --mode READWRITE --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId, fixture: "schema-registry/schema-registry-update-mode.golden"},

		{
			name:    "schema-registry schema create",
			args:    "schema-registry schema create --subject payments --schema=" + schemaPath + " --api-key key --api-secret secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-create.golden",
		},
		{
			name:    "schema-registry schema delete latest",
			args:    "schema-registry schema delete --subject payments --version latest --api-key key --api-secret secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-delete.golden",
		},
		{
			name:    "schema-registry schema delete all",
			args:    "schema-registry schema delete --subject payments --version all --api-key key --api-secret secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-delete-all.golden",
		},
		{
			name:    "schema-registry schema describe --subject payments --version all",
			args:    "schema-registry schema describe --subject payments --version 2 --api-key key --api-secret secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-describe.golden",
		},
		{
			name:    "schema-registry schema describe by id",
			args:    "schema-registry schema describe 10 --api-key key --api-secret secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-describe.golden",
		},

		{
			name:    "schema-registry subject list",
			args:    "schema-registry subject list --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-list.golden",
		},
		{
			name:    "schema-registry subject describe testSubject",
			args:    "schema-registry subject describe testSubject --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-describe.golden",
		},
		{
			name:    "schema-registry subject update compatibility",
			args:    "schema-registry subject update testSubject --compatibility BACKWARD --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-update-compatibility.golden",
		},
		{
			name:    "schema-registry subject update mode",
			args:    "schema-registry subject update testSubject --mode READ --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-update-mode.golden",
		},

		{
			name:    "schema-registry exporter list",
			args:    "schema-registry exporter list --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-list.golden",
		},
		{
			name:    "schema-registry exporter create",
			args:    "schema-registry exporter create myexporter --subjects foo,bar --context-type AUTO --subject-format my-\\${subject} --config-file " + exporterConfigPath + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-create.golden",
		},
		{
			name:    "schema-registry exporter describe",
			args:    "schema-registry exporter describe myexporter " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-describe.golden",
		},
		{
			name:    "schema-registry exporter update",
			args:    "schema-registry exporter update myexporter --subjects foo,bar,test --subject-format my-\\${subject} " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-update.golden",
		},
		{
			name:    "schema-registry exporter delete",
			args:    "schema-registry exporter delete myexporter " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-delete.golden",
		},
		{
			name:    "schema-registry exporter get-status",
			args:    "schema-registry exporter get-status myexporter " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-get-status.golden",
		},
		{
			name:    "schema-registry exporter get-config json",
			args:    "schema-registry exporter get-config myexporter --output json " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-get-config-json.golden",
		},
		{
			name:    "schema-registry exporter get-config yaml",
			args:    "schema-registry exporter get-config myexporter --output yaml " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-get-config-yaml.golden",
		},
		{
			name:    "schema-registry exporter pause",
			args:    "schema-registry exporter pause myexporter " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-pause.golden",
		},
		{
			name:    "schema-registry exporter resume",
			args:    "schema-registry exporter resume myexporter " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-resume.golden",
		},
		{
			name:    "schema-registry exporter reset",
			args:    "schema-registry exporter reset myexporter " + " --api-key=key --api-secret=secret --environment=" + testserver.SRApiEnvId,
			fixture: "schema-registry/schema-registry-exporter-reset.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
