package flink

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *command) newConnectionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink connections.",
		Args:  cobra.NoArgs,
		RunE:  c.connectionList,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionList(cmd *cobra.Command, _ []string) error {
	// TODO
	return nil
}
