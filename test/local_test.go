package test

func (s *CLITestSuite) TestLocalHelpCommands() {
	tests := []CLITest{
		// These should all be equivalent
		{args: "local", fixture: "local-help1.golden"},
		{args: "help local", fixture: "local-help1.golden"},
		{args: "local --help", fixture: "local-help1.golden"},
		// Ideally, this would show subcommand help, but Cobra doesn't send "list" arg to the help command func.
		// So we just show the top level list of commands again. :(
		{args: "help local list", fixture: "local-help1.golden"},
		// But if we call it this way, we can see help for specific local subcommands.
		{args: "local list --help", fixture: "local-help2.golden"},
		// We only have help 2 command levels deep. "local list plugins --help" shows the same as "local list --help"
		{args: "local list plugins --help", fixture: "local-help2.golden"},
	}
	resetConfiguration(s.T(), "confluent")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runConfluentTest(tt, serveMds(s.T(), kafkaAPIURL).URL)
	}
}
