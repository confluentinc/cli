package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *accessPointCommand) newIngressEndpointDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an ingress endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validIngressEndpointArgs),
		RunE:              c.describeIngressEndpoint,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe ingress endpoint "ap-123456".`,
				Code: "confluent network access-point private-link ingress-endpoint describe ap-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *accessPointCommand) describeIngressEndpoint(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	ingressEndpoint, err := c.V2Client.GetAccessPoint(environmentId, args[0])
	if err != nil {
		return err
	}

	return printPrivateLinkIngressEndpointTable(cmd, ingressEndpoint)
}
