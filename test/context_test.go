package test

func (s *CLITestSuite) TestContextCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "config context current", fixture: "context1.golden"},
		{args: "config context list", fixture: "context2.golden"},
		{args: "config context set my-context --kafka-cluster bob", fixture: "context3.golden"},
		{args: "config context list", fixture: "context4.golden"},
		{args: "config context get my-context", fixture: "context5.golden"},
		{args: "config context get other-context", fixture: "context6.golden"},
		{args: "config context list", fixture: "context7.golden"},
		{args: "config context use my-context", fixture: "context8.golden"},
		{args: "config context current", fixture: "context9.golden"},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		s.runTest(tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}
