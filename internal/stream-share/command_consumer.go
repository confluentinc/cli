package streamshare

import (
	"github.com/spf13/cobra"
)

func (c *command) newConsumerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage consumer actions.",
	}

	cmd.AddCommand(c.newConsumerShareCommand())
	cmd.AddCommand(c.newRedeemCommand())

	return cmd
}
