package iam

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/output"
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

func (c *userCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the current user.",
		Args:  cobra.NoArgs,
		RunE:  c.describeOnPrem,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *userCommand) describeOnPrem(cmd *cobra.Command, _ []string) error {
	token := c.Context.GetAuthToken()

	expClaim, err := jwt.GetClaim(token, "exp")
	if err != nil {
		return err
	}
	exp, ok := expClaim.(float64)
	if !ok {
		return fmt.Errorf(errors.MalformedTokenErrorMsg, "exp")
	}

	table := output.NewTable(cmd)
	table.Add(&userOutOnPrem{
		Username:            c.Context.Credential.Username,
		AuthenticationToken: token,
		ExpiresAt:           time.Unix(int64(exp), 0).UTC(),
	})
	return table.Print()
}
