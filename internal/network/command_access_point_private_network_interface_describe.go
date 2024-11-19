package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *accessPointCommand) newPrivateNetworkInterfaceDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a private network interface.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateNetworkInterfaceArgs),
		RunE:              c.privateNetworkInterfaceDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe private network interface "ap-123456".`,
				Code: "confluent network access-point private-network-interface describe ap-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *accessPointCommand) privateNetworkInterfaceDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	privateNetworkInterface, err := c.V2Client.GetAccessPoint(environmentId, args[0])
	if err != nil {
		return err
	}

	return printPrivateNetworkInterfaceTable(cmd, privateNetworkInterface)
}
