package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type invitationCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newInvitationCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invitation",
		Short: "Manage invitations.",
	}

	c := &invitationCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}
