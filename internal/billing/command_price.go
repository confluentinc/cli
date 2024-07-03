package billing

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/spf13/cobra"
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
