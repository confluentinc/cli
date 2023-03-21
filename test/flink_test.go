package test

func (s *CLITestSuite) TestFlinkHelp() {
	tests := []CLITest{
		{args: "flink -h", fixture: "flink/help.golden"},
		{args: "flink statement -h", fixture: "flink/statement/help.golden"},
		{args: "flink statement create -h", fixture: "flink/statement/create-help.golden"},
		{args: "flink statement delete -h", fixture: "flink/statement/delete-help.golden"},
		{args: "flink statement describe -h", fixture: "flink/statement/describe-help.golden"},
		{args: "flink statement list -h", fixture: "flink/statement/list-help.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestFlinkStatement() {
	tests := []CLITest{
		{args: "flink statement create", fixture: "flink/statement/create.golden"},
		{args: "flink statement delete", fixture: "flink/statement/delete.golden"},
		{args: "flink statement describe", fixture: "flink/statement/describe.golden"},
		{args: "flink statement list", fixture: "flink/statement/list.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
