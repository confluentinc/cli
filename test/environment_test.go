package test

func (s *CLITestSuite) TestEnvironment() {
	tests := []CLITest{
		// only login at the begginning so active env is not reset
		// tt.workflow=true so login is not reset
		{args: "environment list", fixture: "environment/1.golden", login: "default"},
		{args: "environment use not-595", fixture: "environment/2.golden"},
		{args: "environment update not-595 --name new-other-name", fixture: "environment/3.golden"},
		{args: "environment list", fixture: "environment/4.golden"},
		{args: "environment list -o json", fixture: "environment/5.golden"},
		{args: "environment list -o yaml", fixture: "environment/6.golden"},
		{args: "environment use non-existent-id", fixture: "environment/7.golden", wantErrCode: 1},
		{args: "environment create saucayyy", fixture: "environment/8.golden"},
		{args: "environment create saucayyy -o json", fixture: "environment/9.golden"},
		{args: "environment create saucayyy -o yaml", fixture: "environment/10.golden"},
		{args: "environment delete not-595", fixture: "environment/11.golden"},
	}

	resetConfiguration(s.T(), "ccloud")
	kafkaURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaURL).URL

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt, loginURL)
	}
}
