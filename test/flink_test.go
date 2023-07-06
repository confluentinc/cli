package test

func (s *CLITestSuite) TestFlinkComputePool() {
	tests := []CLITest{
		{args: "flink compute-pool create my-compute-pool --cloud aws --region us-west-2", fixture: "flink/compute-pool/create.golden"},
		{args: "flink compute-pool delete lfcp-123456 --force", fixture: "flink/compute-pool/delete.golden"},
		{args: "flink compute-pool describe lfcp-123456", fixture: "flink/compute-pool/describe.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list.golden"},
		{args: "flink compute-pool list --region us-west-2", fixture: "flink/compute-pool/list-region.golden"},
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
		{args: "flink statement delete my-statement --force", fixture: "flink/statement/delete.golden"},
		{args: "flink statement list --compute-pool lfcp-123456", fixture: "flink/statement/list.golden"},
		{args: "flink statement describe my-statement", fixture: "flink/statement/describe.golden"},
		{args: "flink statement exception list my-statement", fixture: "flink/statement/exceptions/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkIamBinding() {
	tests := []CLITest{
		{args: "flink iam-binding create --cloud aws --region us-west-2 --identity-pool pool-1234", fixture: "flink/iam-binding/create.golden"},
		{args: "flink iam-binding create --cloud aws --region us-west-2 --identity-pool pool-1234 --environment env-123", fixture: "flink/iam-binding/create-environment.golden"},
		{args: "flink iam-binding delete fiam-123 --force", fixture: "flink/iam-binding/delete.golden"},
		{args: "flink iam-binding list", fixture: "flink/iam-binding/list.golden"},
		{args: "flink iam-binding list --cloud aws", fixture: "flink/iam-binding/list-cloud.golden"},
		{args: "flink iam-binding list --region us-west-1", fixture: "flink/iam-binding/list-region.golden"},
		{args: "flink iam-binding list --identity-pool pool-123", fixture: "flink/iam-binding/list-identity-pool.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
