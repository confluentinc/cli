package test

import (
	"runtime"

	"github.com/confluentinc/cli/internal/pkg/config"
)

func (s *CLITestSuite) TestConfigurationList() {
	cfg := config.AuthenticatedConfigMockWithContextName("context1")
	if err := cfg.Save(); err != nil {
		panic(err)
	}

	test := CLITest{args: "configuration list", fixture: "configuration/list.golden", workflow: true}
	if runtime.GOOS == "windows" {
		test.fixture = "configuration/list-windows.golden"
	}
	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestConfigurationSet() {
	cfg := config.AuthenticatedConfigMockWithContextName("context1")
	if err := cfg.CreateContext("context2", "http://test", "costa", "rica"); err != nil {
		panic(err)
	}
	if err := cfg.Save(); err != nil {
		panic(err)
	}

	tests := []CLITest{
		{args: "configuration set disable_update_check=true", fixture: "configuration/set-one.golden"},
		{args: "configuration set disable_update_check=false disable_plugins=true current_context=context2", fixture: "configuration/set-multiple.golden"},
		{args: "configuration set disable_update_check=yes", fixture: "configuration/set-invalid-1.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}
