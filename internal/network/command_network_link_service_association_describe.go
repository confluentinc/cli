package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *command) newNetworkLinkServiceAssociationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a network link service association.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validNetworkLinkEndpointArgs),
		RunE:              c.networkLinkServiceAssociationDescribe,
	}

	c.addNetworkLinkServiceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network-link-service"))

	return cmd
}

func (c *command) networkLinkServiceAssociationDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	networkLinkService, err := cmd.Flags().GetString("network-link-service")
	if err != nil {
		return err
	}

	association, err := c.V2Client.GetNetworkLinkServiceAssociation(environmentId, networkLinkService, args[0])
	if err != nil {
		return err
	}

	return printNetworkLinkServiceAssociationTable(cmd, association)
}
