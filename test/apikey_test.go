package test

func (s *CLITestSuite) TestAPIKeyCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --cluster bob", useKafka: "bob", fixture: "apikey_create_1.golden"},
		{args: "api-key list", useKafka: "bob", fixture: "apikey_list_1.golden"},
		{args: "api-key list", useKafka: "abc", fixture: "apikey_list_2.golden"},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		s.runTest(tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}
