package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *command) newPrivateLinkAttachmentConnectionDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a private link attachment connection.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.privateLinkAttachmentConnectionDescribe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAttachmentConnectionDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connection, err := c.V2Client.GetPrivateLinkAttachmentConnection(environmentId, args[0])
	if err != nil {
		return err
	}

	return printPrivateLinkAttachmentConnectionTable(cmd, connection)
}
