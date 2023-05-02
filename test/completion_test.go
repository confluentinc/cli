package test

func (s *CLITestSuite) TestCompletion() {
	tests := []CLITest{
		{args: "completion -h", fixture: "completion/help.golden"},
		{args: "completion bash", fixture: "completion/bash.golden"},
		{args: "completion zsh", fixture: "completion/zsh.golden"},
	}

	for _, tt := range tests {
		s.runIntegrationTest(tt)
	}
}
