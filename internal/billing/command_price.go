package billing

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newPriceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "price",
		Short:       "See Confluent Cloud pricing information.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newPriceListCommand())

	return cmd
}
