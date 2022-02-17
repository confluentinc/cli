package test

func (s *CLITestSuite) TestCompletion() {
	tests := []CLITest{
		{fixture: "completion/0.golden", args: "completion -h"},
		{fixture: "completion/1.golden", args: "completion bash"},
		{fixture: "completion/2.golden", args: "completion zsh"},
	}

	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}
