package test

func (s *CLITestSuite) TestVersion() {
	for _, tt := range []CLITest{
		{fixture: "version/version.golden", args: "version"},
		{fixture: "version/version-flag.golden", args: "--version"},
	} {
		tt.regex = true
		s.runConfluentTest(tt)
	}
}
