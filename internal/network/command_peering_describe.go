package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPeeringDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a peering.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPeeringArgs),
		RunE:              c.peeringDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe peering "peer-123456".`,
				Code: "confluent network peering describe peer-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) peeringDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	peering, err := c.V2Client.GetPeering(environmentId, args[0])
	if err != nil {
		return err
	}

	return printPeeringTable(cmd, peering)
}
