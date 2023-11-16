package iam

import (
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *ipFilterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an IP filter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete IP filter "ipf-12345":`,
				Code: "confluent iam ip-filter delete ipf-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) delete(cmd *cobra.Command, args []string) error {
	err := c.V2Client.DeleteIamIpFilter(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.IPFilter, args[0])
	}

	output.Printf(c.Config.EnableColor, "Deleted IP filter \"%s\"\n", args[0])
	return nil
}
