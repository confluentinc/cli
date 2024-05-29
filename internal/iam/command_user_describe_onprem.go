package iam

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/jwt"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *userCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe",
		Short:             "Describe the current user.",
		Args:              cobra.NoArgs,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describeOnPrem,
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
		ExpiresAt:           time.Unix(0, int64(exp)).UTC(),
	})
	return table.Print()
}
