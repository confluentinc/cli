package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newTransitGatewayAttachmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a transit gateway attachment.",
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validTransitGatewayAttachmentArgs),
		Args:              cobra.ExactArgs(1),
		RunE:              c.transitGatewayAttachmentDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe transit gateway attachment "tgwa-123456".`,
				Code: "confluent network transit-gateway-attachment describe tgwa-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) transitGatewayAttachmentDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	attachment, err := c.V2Client.GetTransitGatewayAttachment(environmentId, args[0])
	if err != nil {
		return err
	}

	return printTransitGatewayAttachmentTable(cmd, attachment)
}
