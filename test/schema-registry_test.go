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
		{Args: "schema-registry --help", Fixture: "schema-registry/schema-registry-help.golden"},
		{Args: "schema-registry cluster --help", Fixture: "schema-registry/schema-registry-cluster-help.golden"},
		{Args: "schema-registry cluster enable --cloud gcp --geo us -o json", Fixture: "schema-registry/schema-registry-enable-json.golden"},
		{Args: "schema-registry cluster enable --cloud gcp --geo us -o yaml", Fixture: "schema-registry/schema-registry-enable-yaml.golden"},
		{Args: "schema-registry cluster enable --cloud gcp --geo us", Fixture: "schema-registry/schema-registry-enable.golden"},
		{Args: "schema-registry schema --help", Fixture: "schema-registry/schema-registry-schema-help.golden"},
		{Args: "schema-registry subject --help", Fixture: "schema-registry/schema-registry-subject-help.golden"},

		{Args: "schema-registry cluster describe", Fixture: "schema-registry/schema-registry-describe.golden"},
		{Args: "schema-registry cluster update --environment=" + test_server.SRApiEnvId, Fixture: "schema-registry/schema-registry-update-missing-flags.golden", WantErrCode: 1},
		{Args: "schema-registry cluster update --compatibility BACKWARD --environment=" + test_server.SRApiEnvId, PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("key\nsecret\n"))}, Fixture: "schema-registry/schema-registry-update-compatibility.golden"},
		{Args: "schema-registry cluster update --mode READWRITE --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId, Fixture: "schema-registry/schema-registry-update-mode.golden"},

		{
			Name:    "schema-registry schema create",
			Args:    "schema-registry schema create --subject payments --schema=" + schemaPath + " --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-schema-create.golden",
		},
		{
			Name:    "schema-registry schema delete latest",
			Args:    "schema-registry schema delete --subject payments --version latest --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-schema-delete.golden",
		},
		{
			Name:    "schema-registry schema delete all",
			Args:    "schema-registry schema delete --subject payments --version all --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-schema-delete-all.golden",
		},
		{
			Name:    "schema-registry schema describe --subject payments --version all",
			Args:    "schema-registry schema describe --subject payments --version 2 --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-schema-describe.golden",
		},
		{
			Name:    "schema-registry schema describe by id",
			Args:    "schema-registry schema describe 10 --api-key key --api-secret secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-schema-describe.golden",
		},

		{
			Name:    "schema-registry subject list",
			Args:    "schema-registry subject list --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-subject-list.golden",
		},
		{
			Name:    "schema-registry subject describe testSubject",
			Args:    "schema-registry subject describe testSubject --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-subject-describe.golden",
		},
		{
			Name:    "schema-registry subject update compatibility",
			Args:    "schema-registry subject update testSubject --compatibility BACKWARD --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-subject-update-compatibility.golden",
		},
		{
			Name:    "schema-registry subject update mode",
			Args:    "schema-registry subject update testSubject --mode READ --api-key=key --api-secret=secret --environment=" + test_server.SRApiEnvId,
			Fixture: "schema-registry/schema-registry-subject-update-mode.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
