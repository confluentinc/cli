package test

func (s *CLITestSuite) TestFlinkHelp() {
	tests := []CLITest{
		{args: "flink -h", fixture: "flink/help.golden"},
		{args: "flink sql-statement -h", fixture: "flink/sql-statement/help.golden"},
		{args: "flink sql-statement create -h", fixture: "flink/sql-statement/create-help.golden"},
		{args: "flink sql-statement delete -h", fixture: "flink/sql-statement/delete-help.golden"},
		{args: "flink sql-statement describe -h", fixture: "flink/sql-statement/describe-help.golden"},
		{args: "flink sql-statement list -h", fixture: "flink/sql-statement/list-help.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestFlinkSqlStatement() {
	tests := []CLITest{
		{args: "flink sql-statement create", fixture: "flink/sql-statement/create.golden"},
		{args: "flink sql-statement delete", fixture: "flink/sql-statement/delete.golden"},
		{args: "flink sql-statement describe", fixture: "flink/sql-statement/describe.golden"},
		{args: "flink sql-statement list", fixture: "flink/sql-statement/list.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
