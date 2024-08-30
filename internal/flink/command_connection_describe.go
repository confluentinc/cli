package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *command) newConnectionDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [name]",
		Short:             "Describe a Flink connection.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgs), // TODO: update this to connection
		RunE:              c.connectionDescribe,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionDescribe(cmd *cobra.Command, args []string) error {
	// TODO
	return nil
}
