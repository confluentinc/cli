package iam

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type userOutOnPrem struct {
	Username            string    `human:"Username" serialized:"username"`
	AuthenticationToken string    `human:"Authentication Token" serialized:"authentication_token"`
	ExpiresAt           time.Time `human:"Expires At" serialized:"expires_at"`
}

// newUserCommandOnPrem builds the on-prem variant of `confluent iam user`, which
// exposes a single `describe` reading the local auth token. The parent command.go
// selects this or the (generated) cloud newUserCommand by login mode.
func newUserCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "user",
		Short:       "Manage users.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &userCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}

	cmd.AddCommand(c.newDescribeCommandOnPrem())

	return cmd
}
