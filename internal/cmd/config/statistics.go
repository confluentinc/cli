package config

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const (
	statisticsCommandName = "statistics"
)

type statisticsCommand struct {
	*pcmd.AuthenticatedCLICommand
	contextCommand *contextCommand
}

func NewTracking(contextCmd *contextCommand) *cobra.Command {
	var cliCmd *pcmd.AuthenticatedCLICommand
	if contextCmd.cliName == "confluent" {
		cliCmd = pcmd.NewAuthenticatedWithMDSCLICommand(
			&cobra.Command{
				Use:   statisticsCommandName,
				Short: "Disable or enable usage statistics.",
			}, contextCmd.prerunner)
	} else {
		cliCmd = pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   statisticsCommandName,
				Short: "Disable or enable usage statistics.",
			}, contextCmd.prerunner)
	}
	cmd := &statisticsCommand{
		AuthenticatedCLICommand: cliCmd,
		contextCommand:          contextCmd,
	}
	cmd.init()
	return cmd.Command
}

func (c *statisticsCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "disable",
		Short: "Disable usage statistics.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.disable),
	})

	c.AddCommand(&cobra.Command{
		Use:   "enable",
		Short: "Enable usage statistics.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.enable),
	})
}

func (c *statisticsCommand) disable(cmd *cobra.Command, _ []string) error {
	ctx := c.Config.Config.Context()
	ctx.DisableTracking = true
	return c.Config.Save()
}

func (c *statisticsCommand) enable(cmd *cobra.Command, _ []string) error {
	ctx := c.Config.Config.Context()
	ctx.DisableTracking = false
	return c.Config.Save()
}
