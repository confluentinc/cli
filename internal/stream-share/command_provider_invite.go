package streamshare

import (
	"github.com/spf13/cobra"
)

func (c *command) newInviteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Manage invites.",
		Args:  cobra.ExactArgs(1),
	}

	cmd.AddCommand(c.newCreateEmailInviteCommand())
	cmd.AddCommand(c.newResendEmailInviteCommand())

	return cmd
}
