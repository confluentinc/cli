package test

func (s *CLITestSuite) TestVersion() {
	for _, test := range []CLITest{
		{fixture: "version/version.golden", args: "version"},
		{fixture: "version/version-flag.golden", args: "--version"},
	} {
		test.regex = true
		s.runIntegrationTest(test)
	}
}
