package config

/*
 * Using ccloud config for quickstart example
 * ccloud <resource> [subresource] <standard-verb> [args]
 * <resource> = config in this case
 */

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

// Additional wrapper around a command for testing purposes?
type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	analytics analytics.Client
	config    *v3.Config
}

// New returns the Cobra command for `config`.
// This specifies the command created for 'config' (<resource> <no-verb>), it is called/registered by the top level (cmd/command.go)
func New(prerunner pcmd.PreRunner, analytics analytics.Client, config *v3.Config) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "config",
			Short: "Modify the CLI configuration.",
		}, prerunner)
	cmd := &command{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		analytics:  analytics,
		config:     config,
	}
	cmd.init()
	return cmd.Command
}

// CLI codebase convention, specify subresources and verbs of this resource here
func (c *command) init() {
	c.AddCommand(NewContext(c.prerunner, c.analytics))
	c.AddCommand(NewFile(c.prerunner, c.config))
}
