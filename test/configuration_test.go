package test

import (
	"runtime"
)

func (s *CLITestSuite) TestConfigurationDescribe() {
	tests := []CLITest{
		{args: "configuration describe disable_plugins", fixture: "configuration/describe-1.golden"},
		{args: "configuration describe contexts", fixture: "configuration/describe-invalid-1.golden", exitCode: 1},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConfigurationList() {
	test := CLITest{args: "configuration list", fixture: "configuration/list.golden"}
	if runtime.GOOS == "windows" {
		test.fixture = "configuration/list-windows.golden"
	}
	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestConfigurationUpdate() {
	tests := []CLITest{
		{args: "configuration update disable_update_check true", fixture: "configuration/update.golden"},
		{args: "configuration update disable_update_check yes", fixture: "configuration/update-invalid-1.golden", exitCode: 1},
		{args: "configuration update current_context new-context", fixture: "configuration/update-invalid-2.golden", exitCode: 1},
		{args: "configuration update platforms nil", fixture: "configuration/update-invalid-3.golden", exitCode: 1},
		{args: "configuration update disable_feature_flags true", fixture: "configuration/update-prompt-cancel.golden", input: "n\n"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConfiguration_Autocomplete() {
	test := CLITest{args: `__complete configuration update ""`, fixture: "configuration/update-autocomplete.golden"}
	if runtime.GOOS == "windows" {
		test.fixture = "configuration/update-autocomplete-windows.golden"
	}
	s.runIntegrationTest(test)
}
