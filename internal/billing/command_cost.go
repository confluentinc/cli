package billing

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newCostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cost",
		Short:       "List Confluent Cloud billing costs.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newCostListCommand())

	return cmd
}
