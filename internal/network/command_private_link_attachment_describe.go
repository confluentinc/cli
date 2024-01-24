package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAttachmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a private link attachment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAttachmentArgs),
		RunE:              c.privateLinkAttachmentDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe private link attachment "platt-123456".`,
				Code: "confluent network private-link attachment describe platt-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAttachmentDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	attachment, err := c.V2Client.GetPrivateLinkAttachment(environmentId, args[0])
	if err != nil {
		return err
	}

	return printPrivateLinkAttachmentTable(cmd, attachment)
}
