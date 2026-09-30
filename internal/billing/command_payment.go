package billing

import (
	"github.com/spf13/cobra"
)

func (c *command) newPaymentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "Manage payment method.",
	}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}
