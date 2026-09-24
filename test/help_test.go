package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/internal"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/version"
)

// Cloud and on-prem are separate methods so they can run in different CI shards.
func (s *CLITestSuite) TestHelpCloud() {
	s.testHelpForPlatform("cloud", "https://confluent.cloud")
}

func (s *CLITestSuite) TestHelpOnPrem() {
	s.testHelpForPlatform("onprem", "https://example.com")
}

func (s *CLITestSuite) testHelpForPlatform(login, platformName string) {
	cfg := &config.Config{
		CurrentContext:      login,
		Contexts:            map[string]*config.Context{login: {PlatformName: platformName}},
		Version:             new(version.Version),
		IsTest:              true,
		DisableFeatureFlags: true,
	}

	s.testHelp(internal.NewConfluentCommand(cfg), login)
}

func (s *CLITestSuite) testHelp(cmd *cobra.Command, login string) {
	path := strings.Split(cmd.CommandPath(), " ")[1:]
	args := append(path, "--help")

	file := "help.golden"
	if login != "cloud" {
		file = fmt.Sprintf("help-%s.golden", login)
	}

	if cmd.HasSubCommands() || len(path) == 1 {
		path = append(path, file)
	} else {
		path[len(path)-1] += "-" + file
	}

	test := CLITest{
		args:    strings.Join(args, " "),
		fixture: filepath.Join(path...),
		login:   login,
	}

	if strings.Contains(test.args, "services kafka produce") || strings.Contains(test.args, "services kafka consume") {
		test.regex = true
	}

	if cmd.IsAvailableCommand() {
		s.runIntegrationTest(test)
	} else {
		_ = os.RemoveAll(test.fixture)
	}

	for _, subcommand := range cmd.Commands() {
		s.testHelp(subcommand, login)
	}
}

func (s *CLITestSuite) TestHelp_AllFormats() {
	tests := []CLITest{
		{args: ""},
		{args: "-h"},
		{args: "help"},
	}

	for _, test := range tests {
		test.fixture = "help.golden"
		test.login = "cloud"
		s.runIntegrationTest(test)

		test.fixture = "help-onprem.golden"
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}
