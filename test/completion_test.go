package test

func (s *CLITestSuite) TestCompletion() {
	tests := []CLITest{
		{args: "completion bash", fixture: "completion/bash.golden"},
		{args: "completion zsh", fixture: "completion/zsh.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}
