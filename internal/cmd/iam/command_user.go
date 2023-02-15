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

	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(newInvitationCommand(prerunner))
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}
