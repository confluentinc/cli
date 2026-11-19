package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/internal"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/version"
)

func (s *CLITestSuite) TestHelp() {
	configurations := []*config.Config{
		{
			CurrentContext: "cloud",
			Contexts:       map[string]*config.Context{"cloud": {PlatformName: "https://confluent.cloud"}},
		},
		{
			CurrentContext: "onprem",
			Contexts:       map[string]*config.Context{"onprem": {PlatformName: "https://example.com"}},
		},
	}

	for _, cfg := range configurations {
		cfg.Version = new(version.Version)
		cfg.IsTest = true
		cfg.DisableFeatureFlags = true

		s.testHelp(internal.NewConfluentCommand(cfg), cfg.CurrentContext)
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

func (s *CLITestSuite) TestHelp_Streams() {
	tests := []struct {
		args       string
		exitCode   int
		wantStdout bool // true if help output is intentional, false otherwise
	}{
		{args: "--help", wantStdout: true},
		{args: "context", wantStdout: true}, // parent command (considered intentional)
		{args: "context use", exitCode: 1},  // missing required argument
	}

	for _, test := range tests {
		s.T().Run(test.args, func(t *testing.T) {
			resetConfiguration(t, false)

			stdout, stderr := runCommandSeparate(t, testBin, nil, test.args, test.exitCode, "")

			if test.wantStdout {
				require.Contains(t, stdout, "Usage:")
				require.Empty(t, stderr)
			} else {
				require.Empty(t, stdout)
				require.Contains(t, stderr, "Error:")
				require.Contains(t, stderr, "Usage:")
			}
		})
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
