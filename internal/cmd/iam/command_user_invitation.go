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

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}
