package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type inviteCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newInviteCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Manage invites.",
		Args:  cobra.ExactArgs(1),
	}

	c := &inviteCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newResendCommand())

	return c.Command
}
