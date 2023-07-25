package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (s *CLITestSuite) TestHelp() {
	configurations := []*v1.Config{
		{
			CurrentContext: "cloud",
			Contexts:       map[string]*v1.Context{"cloud": {PlatformName: "https://confluent.cloud"}},
		},
		{
			CurrentContext: "onprem",
			Contexts:       map[string]*v1.Context{"onprem": {PlatformName: "https://example.com"}},
		},
	}

	for _, cfg := range configurations {
		cfg.Version = new(version.Version)
		cfg.IsTest = true
		cfg.DisableFeatureFlags = true

		s.testHelp(pcmd.NewConfluentCommand(cfg), cfg.CurrentContext)
	}
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
