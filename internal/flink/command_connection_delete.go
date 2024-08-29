package flink

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *command) newConnectionDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <name-1> [name-2] ... [name-n]",
		Short:             "Delete one or more Flink connections.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgsMultiple), // TODO: update this
		RunE:              c.connectionDelete,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionDelete(cmd *cobra.Command, args []string) error {
	// TODO
	return nil
}
