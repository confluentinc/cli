package test

import "runtime"

func (s *CLITestSuite) TestHelp_NoContext() {
	for _, tt := range []CLITest{
		{args: ""},
		{args: "help"},
		{args: "-h"},
		{args: "--help"},
	} {
		if runtime.GOOS == "windows" {
			tt.fixture = "help/help-no-context-windows.golden"
		} else {
			tt.fixture = "help/help-no-context.golden"
		}

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
		if runtime.GOOS == "windows" {
			tt.fixture = "help/help-cloud-windows.golden"
		} else {
			tt.fixture = "help/help-cloud.golden"
		}

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
		if runtime.GOOS == "windows" {
			tt.fixture = "help/help-onprem-windows.golden"
		} else {
			tt.fixture = "help/help-onprem.golden"
		}

		tt.login = "default"
		s.runConfluentTest(tt)
	}
}
