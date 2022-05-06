package test

import testserver "github.com/confluentinc/cli/test/test-server"

func (s *CLITestSuite) TestStreamGovernance() {
	tests := []CLITest{
		{args: "stream-governance --help", fixture: "stream-governance/help.golden"},
		{args: "stream-governance describe  --environment=" + testserver.SRApiEnvId, fixture: "stream-governance/describe.golden"},
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
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
