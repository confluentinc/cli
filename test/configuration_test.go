package test

import (
	"runtime"
)

func (s *CLITestSuite) TestConfigurationView() {
	test := CLITest{args: "configuration view", fixture: "configuration/view.golden"}
	if runtime.GOOS == "windows" {
		test.fixture = "configuration/view-windows.golden"
	}
	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestConfigurationSet() {
	tests := []CLITest{
		{args: "configuration set disable_update_check true", fixture: "configuration/set.golden"},
		{args: "configuration set disable_update_check yes", fixture: "configuration/set-invalid-1.golden", exitCode: 1},
		{args: "configuration set current_context new-context", fixture: "configuration/set-invalid-2.golden", exitCode: 1},
		{args: "configuration set platforms nil", fixture: "configuration/set-invalid-3.golden", exitCode: 1},
		{args: "configuration set disable_feature_flags true", fixture: "configuration/set-prompt-cancel.golden", input: "n\n"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConfiguration_Autocomplete() {
	test := CLITest{args: `__complete configuration set ""`, fixture: "configuration/set-autocomplete.golden"}
	if runtime.GOOS == "windows" {
		test.fixture = "configuration/set-autocomplete-windows.golden"
	}
	s.runIntegrationTest(test)
}
