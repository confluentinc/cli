package test

import (
	"github.com/confluentinc/bincover"
	test_server "github.com/confluentinc/cli/test/test-server"
	"strings"
)

func (s *CLITestSuite) TestSchemaRegistry() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "schema-registry --help", fixture: "schema-registry/schema-registry-help.golden"},
		{args: "schema-registry cluster --help", fixture: "schema-registry/schema-registry-cluster-help.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o json", fixture: "schema-registry/schema-registry-enable-json.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o yaml", fixture: "schema-registry/schema-registry-enable-yaml.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us", fixture: "schema-registry/schema-registry-enable.golden"},
		{args: "schema-registry schema --help", fixture: "schema-registry/schema-registry-schema-help.golden"},
		{args: "schema-registry subject --help", fixture: "schema-registry/schema-registry-subject-help.golden"},

		{args: "schema-registry cluster describe", fixture: "schema-registry/schema-registry-describe.golden"},
		{args: "schema-registry cluster update --environment="+test_server.SRUpdateEnvId, fixture: "schema-registry/schema-registry-update-missing-flags.golden", wantErrCode: 1},
		{args: "schema-registry cluster update --compatibility BACKWARD --environment="+test_server.SRUpdateEnvId, preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("key\nsecret\n"))}, fixture: "schema-registry/schema-registry-update-compatibility.golden"},
		{args: "schema-registry cluster update --mode READWRITE --api-key=key --api-secret=secret --environment="+test_server.SRUpdateEnvId, fixture: "schema-registry/schema-registry-update-mode.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
