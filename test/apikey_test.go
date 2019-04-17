package test

func (s *CLITestSuite) TestAPIKeyCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --cluster bob", useKafka: "bob", fixture: "apikey3.golden", wantErrCode: 0},
		{args: "api-key list", useKafka: "bob", fixture: "apikey1.golden", wantErrCode: 0},
		{args: "api-key list", useKafka: "abc", fixture: "apikey2.golden", wantErrCode: 0},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		runTest(s.T(), tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}
