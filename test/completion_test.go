package test

func (s *CLITestSuite) TestCompletion() {
	tests := []CLITest{
		{args: "completion -h", fixture: "completion/0.golden"},
		{args: "completion bash", contains: "# bash completion for confluent"},
		{args: "completion zsh", contains: "# zsh completion for confluent"},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}
