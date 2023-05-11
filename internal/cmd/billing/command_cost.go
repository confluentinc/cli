package billing

import (
	"github.com/spf13/cobra"
)

func (c *command) newCostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "List Confluent Cloud billing costs.",
	}

	cmd.AddCommand(c.newCostListCommand())

	return cmd
}
