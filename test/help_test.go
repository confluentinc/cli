package test

import "runtime"

var helpTests = []CLITest{
	{args: ""},
	{args: "help"},
	{args: "-h"},
	{args: "--help"},
}

func (s *CLITestSuite) TestHelp_NoContext() {
	for _, tt := range helpTests {
		if runtime.GOOS == "windows" {
			tt.fixture = "help/no-context-windows.golden"
		} else {
			tt.fixture = "help/no-context.golden"
		}

		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestHelp_Cloud() {
	for _, tt := range helpTests {
		if runtime.GOOS == "windows" {
			tt.fixture = "help/cloud-windows.golden"
		} else {
			tt.fixture = "help/cloud.golden"
		}

		tt.login = "default"
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestHelp_OnPrem() {
	for _, tt := range helpTests {
		if runtime.GOOS == "windows" {
			tt.fixture = "help/onprem-windows.golden"
		} else {
			tt.fixture = "help/onprem.golden"
		}

		tt.login = "default"
		s.runConfluentTest(tt)
	}
}
