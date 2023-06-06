package test

import (
	"fmt"
	"path/filepath"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (s *CLITestSuite) TestHelp_AllFormats() {
	tests := []CLITest{
		{args: ""},
		{args: "--help"},
		{args: "-h"},
		{args: "help"},
	}

	for _, test := range tests {
		test.fixture = "help/no-context.golden"
		test.login = ""
		s.runIntegrationTest(test)

		test.fixture = "help/cloud.golden"
		test.login = "cloud"
		s.runIntegrationTest(test)

		test.fixture = "help/onprem.golden"
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestHelp_Cloud() {
	cfg := &v1.Config{
		CurrentContext:      "Cloud",
		Contexts:            map[string]*v1.Context{"Cloud": {PlatformName: "https://confluent.cloud"}},
		IsTest:              true,
		Version:             new(version.Version),
		DisableFeatureFlags: true,
	}

	cmd := pcmd.NewConfluentCommand(cfg)
	for _, subcommand := range cmd.Commands() {
		s.testHelp(subcommand, "cloud")
	}
}

func (s *CLITestSuite) TestHelp_OnPrem() {
	cfg := &v1.Config{
		CurrentContext:      "On-Prem",
		Contexts:            map[string]*v1.Context{"On-Prem": {PlatformName: "https://example.com"}},
		IsTest:              true,
		Version:             new(version.Version),
		DisableFeatureFlags: true,
	}

	cmd := pcmd.NewConfluentCommand(cfg)
	for _, subcommand := range cmd.Commands() {
		s.testHelp(subcommand, "onprem")
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

	s.runIntegrationTest(test)

	for _, subcommand := range cmd.Commands() {
		s.testHelp(subcommand, login)
	}
}
