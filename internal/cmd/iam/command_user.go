package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var authMethodFormats = map[string]string{
	"AUTH_TYPE_LOCAL":   "Username/Password",
	"AUTH_TYPE_SSO":     "SSO",
	"AUTH_TYPE_UNKNOWN": "Unknown",
}

type userCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type userOut struct {
	Id                   string `human:"ID" serialized:"id"`
	Name                 string `human:"Name" serialized:"name"`
	Email                string `human:"Email" serialized:"email"`
	AuthenticationMethod string `human:"Authentication Method" serialized:"authentication_method"`
}

func newUserCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "user",
		Short:       "Manage users.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &userCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(newInvitationCommand(prerunner))
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *userCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteUsers(c.V2Client)
}
