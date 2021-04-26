package test

import (
	"strings"

	"github.com/confluentinc/bincover"

	test_server "github.com/confluentinc/cli/test/test-server"
)

func (s *CLITestSuite) TestSchemaRegistry() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	schemaPath := GetInputFixturePath(s.T(), "schema", "schema-example.json")

	tests := []CLITest{
		{args: "schema-registry --help", fixture: "schema-registry/schema-registry-help.golden"},
		{args: "schema-registry cluster --help", fixture: "schema-registry/schema-registry-cluster-help.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o json", fixture: "schema-registry/schema-registry-enable-json.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o yaml", fixture: "schema-registry/schema-registry-enable-yaml.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us", fixture: "schema-registry/schema-registry-enable.golden"},
		{args: "schema-registry schema --help", fixture: "schema-registry/schema-registry-schema-help.golden"},
		{args: "schema-registry subject --help", fixture: "schema-registry/schema-registry-subject-help.golden"},

		{args: "schema-registry cluster describe", fixture: "schema-registry/schema-registry-describe.golden"},
		{args: "schema-registry cluster update --environment=" + test_server.SRApiEnvId, fixture: "schema-registry/schema-registry-update-missing-flags.golden", wantErrCode: 1},
		{args: "schema-registry cluster update --compatibility BACKWARD --environment=" + test_server.SRApiEnvId, preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("key\nsecret\n"))}, fixture: "schema-registry/schema-registry-update-compatibility.golden"},
		{args: "schema-registry cluster update --mode READWRITE --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId, fixture: "schema-registry/schema-registry-update-mode.golden"},

		{
			name:    "schema-registry schema create",
			args:    "schema-registry schema create --subject payments --schema=" + schemaPath + " --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-create.golden",
		},
		{
			name:    "schema-registry schema delete latest",
			args:    "schema-registry schema delete --subject payments --version latest --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-delete.golden",
		},
		{
			name:    "schema-registry schema delete all",
			args:    "schema-registry schema delete --subject payments --version all --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-delete-all.golden",
		},
		{
			name:    "schema-registry schema describe --subject payments --version all",
			args:    "schema-registry schema describe --subject payments --version 2 --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-describe.golden",
		},
		{
			name:    "schema-registry schema describe by id",
			args:    "schema-registry schema describe 10 --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-schema-describe.golden",
		},

		{
			name:    "schema-registry subject list",
			args:    "schema-registry subject list --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-list.golden",
		},
		{
			name:    "schema-registry subject describe testSubject",
			args:    "schema-registry subject describe testSubject --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-describe.golden",
		},
		{
			name:    "schema-registry subject update compatibility",
			args:    "schema-registry subject update testSubject --compatibility BACKWARD --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-update-compatibility.golden",
		},
		{
			name:    "schema-registry subject update mode",
			args:    "schema-registry subject update testSubject --mode READ --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			fixture: "schema-registry/schema-registry-subject-update-mode.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
