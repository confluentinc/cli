package iam

import (
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *ipGroupCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an IP group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete IP group "ipg-12345":`,
				Code: "confluent iam ip-group delete ipg-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *ipGroupCommand) delete(cmd *cobra.Command, args []string) error {
	err := c.V2Client.DeleteIamIPGroup(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.IPGroup, args[0])
	}

	output.Printf(false, "Successfully deleted IP group: %s\n", args[0])
	return nil
}
