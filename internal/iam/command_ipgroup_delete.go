package iam

import (
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *ipGroupCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an IP Group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete IP Group "ipg-12345":`,
				Code: "confluent iam pool delete ipg-12345",
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)
	return cmd
}

func (c *ipGroupCommand) delete(cmd *cobra.Command, args []string) error {

	err := c.V2Client.DeleteIamIPGroup(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.IPGroup, args[0])
	}

	fmt.Printf("Successfully deleted IP Group: %s\n", args[0])
	return err
}
