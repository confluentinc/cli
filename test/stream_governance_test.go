package test

import (
	"github.com/confluentinc/bincover"
	testserver "github.com/confluentinc/cli/test/test-server"
	"strings"
)

func (s *CLITestSuite) TestStreamGovernance() {
	tests := []CLITest{
		{
			args:    "stream-governance --help",
			fixture: "stream-governance/help.golden"},
		{
			args:    "stream-governance describe  --environment=" + testserver.SRApiEnvId,
			fixture: "stream-governance/describe.golden"},
		{
			args:    "stream-governance enable --cloud aws --region us-east-2 --package advanced --environment=" + testserver.SRApiEnvId,
			fixture: "stream-governance/enable-human.golden",
		},
		{
			args:    "stream-governance enable --cloud aws --region us-east-2 --package advanced -o json --environment=" + testserver.SRApiEnvId,
			fixture: "stream-governance/enable-json.golden",
		},
		{
			args:    "stream-governance enable --cloud aws --region us-east-2 --package advanced -o yaml --environment=" + testserver.SRApiEnvId,
			fixture: "stream-governance/enable-yaml.golden",
		},
		{
			args:        "stream-governance enable --region us-east-2 --package advanced --environment=" + testserver.SRApiEnvId,
			fixture:     "stream-governance/enable-missing-flag.golden",
			wantErrCode: 1,
		},
		{
			args:        "stream-governance enable --cloud invalid-cloud --region us-east-2 --package advanced --environment=" + testserver.SRApiEnvId,
			fixture:     "stream-governance/enable-invalid-cloud.golden",
			wantErrCode: 1,
		},
		{
			args:        "stream-governance enable --cloud aws --region invalid-region --package advanced --environment=" + testserver.SRApiEnvId,
			fixture:     "stream-governance/enable-invalid-region.golden",
			wantErrCode: 1,
		},
		{
			args:    "stream-governance upgrade --package advanced --environment=" + testserver.SRApiEnvId,
			fixture: "stream-governance/upgrade.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("y\n"))},
			args:        "stream-governance delete --environment=" + testserver.SRApiEnvId,
			fixture:     "stream-governance/delete.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("n\n"))},
			args:        "stream-governance delete --environment=" + testserver.SRApiEnvId,
			fixture:     "stream-governance/delete-terminated.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
