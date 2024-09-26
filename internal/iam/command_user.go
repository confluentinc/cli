package iam

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

const badResourceIdErrorMsg = `failed parsing resource ID: missing prefix "%s-" is required`

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

type userOutOnPrem struct {
	Username            string    `human:"Username" serialized:"username"`
	AuthenticationToken string    `human:"Authentication Token" serialized:"authentication_token"`
	ExpiresAt           time.Time `human:"Expires At" serialized:"expires_at"`
}

func newUserCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "user",
		Short:       "Manage users.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	c := &userCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(newInvitationCommand(prerunner))
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newUpdateCommand())
	} else {
		cmd.AddCommand(c.newDescribeCommandOnPrem())
	}

	return cmd
}

func (c *userCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *userCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteUsers(c.V2Client)
}
