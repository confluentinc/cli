package test

func (s *CLITestSuite) TestEnvironmentCommands() {
	kafkaAPIURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaAPIURL).URL
	tests := []CLITest{
		{args: "environment list", fixture: "environment1.golden", wantErrCode: 0},
		{args: "environment use not-595", fixture: "environment2.golden", wantErrCode: 0},
		{args: "environment update not-595 --name new-other-name", fixture: "environment10.golden", wantErrCode: 0},
		{args: "environment list", fixture: "environment3.golden", wantErrCode: 0},
		{args: "environment list -o json", fixture: "environment4.golden", wantErrCode: 0},
		{args: "environment list -o yaml", fixture: "environment5.golden", wantErrCode: 0},
		{args: "environment use non-existent-id", fixture: "environment6.golden", wantErrCode: 1},
		{args: "environment create saucayyy", fixture: "environment7.golden", wantErrCode: 0},
		{args: "environment create saucayyy -o json", fixture: "environment8.golden", wantErrCode: 0},
		{args: "environment create saucayyy -o yaml", fixture: "environment9.golden", wantErrCode: 0},
		{args: "environment delete not-595", fixture: "environment11.golden", wantErrCode: 0},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		s.runCcloudTest(tt, loginURL)
	}
}
