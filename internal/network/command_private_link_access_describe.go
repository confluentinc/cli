package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAccessDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a private link access.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAccessArgs),
		RunE:              c.privateLinkAccessDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe private link access "pla-123456".`,
				Code: "confluent network private-link access describe pla-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAccessDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	access, err := c.V2Client.GetPrivateLinkAccess(environmentId, args[0])
	if err != nil {
		return err
	}

	return printPrivateLinkAccessTable(cmd, access)
}
