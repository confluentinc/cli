package test

import "runtime"

func (s *CLITestSuite) TestHelp_NoContext() {
	for _, tt := range []CLITest{
		{args: ""},
		{args: "help"},
		{args: "-h"},
		{args: "--help"},
	} {
		tt.fixture = "help/help-no-context.golden"
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestHelp_Cloud() {
	for _, tt := range []CLITest{
		{args: ""},
		{args: "help"},
		{args: "-h"},
		{args: "--help"},
	} {
		tt.fixture = "help/help-cloud.golden"
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestHelp_OnPrem() {
	for _, tt := range []CLITest{
		{args: ""},
		{args: "help"},
		{args: "-h"},
		{args: "--help"},
	} {
		tt.fixture = "help/help-onprem.golden"
		if runtime.GOOS == "windows" {
			tt.fixture = "help/help-onprem-windows.golden"
		}
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}
