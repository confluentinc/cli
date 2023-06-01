package test

func (s *CLITestSuite) TestFlinkHelp() {
	tests := []CLITest{
		{args: "flink -h", fixture: "flink/help.golden"},
		{args: "flink compute-pool -h", fixture: "flink/compute-pool/help.golden"},
		{args: "flink compute-pool create -h", fixture: "flink/compute-pool/create-help.golden"},
		{args: "flink compute-pool delete -h", fixture: "flink/compute-pool/delete-help.golden"},
		{args: "flink compute-pool describe -h", fixture: "flink/compute-pool/describe-help.golden"},
		{args: "flink compute-pool list -h", fixture: "flink/compute-pool/list-help.golden"},
		{args: "flink compute-pool update -h", fixture: "flink/compute-pool/update-help.golden"},
		{args: "flink compute-pool use -h", fixture: "flink/compute-pool/use-help.golden"},
		{args: "flink region -h", fixture: "flink/region/help.golden"},
		{args: "flink region list -h", fixture: "flink/region/list-help.golden"},
		{args: "flink statement -h", fixture: "flink/statement/help.golden"},
		{args: "flink statement delete -h", fixture: "flink/statement/delete-help.golden"},
		{args: "flink statement list -h", fixture: "flink/statement/list-help.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePool() {
	tests := []CLITest{
		{args: "flink compute-pool create my-compute-pool --cloud aws --region us-west-2", fixture: "flink/compute-pool/create.golden"},
		{args: "flink compute-pool delete lfcp-123456 --force", fixture: "flink/compute-pool/delete.golden"},
		{args: "flink compute-pool describe lfcp-123456", fixture: "flink/compute-pool/describe.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list.golden"},
		{args: "flink compute-pool update lfcp-123456 --cfu 2", fixture: "flink/compute-pool/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUse() {
	tests := []CLITest{
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use.golden"},
		{args: "flink compute-pool describe", fixture: "flink/compute-pool/describe-after-use.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list-after-use.golden"},
		{args: "flink compute-pool update --cfu 2", fixture: "flink/compute-pool/update-after-use.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkRegion() {
	tests := []CLITest{
		{args: "flink region list", fixture: "flink/region/list.golden"},
		{args: "flink region list -o json", fixture: "flink/region/list-json.golden"},
		{args: "flink region list --cloud aws", fixture: "flink/region/list-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatement() {
	tests := []CLITest{
		{args: "flink statement delete my-statement --compute-pool lfcp-123456 --force", fixture: "flink/statement/delete.golden"},
		{args: "flink statement list --compute-pool lfcp-123456", fixture: "flink/statement/list.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
