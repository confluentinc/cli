package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c userCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a user.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c userCommand) describe(cmd *cobra.Command, args []string) error {
	if resource.LookupType(args[0]) != resource.User {
		return fmt.Errorf(errors.BadResourceIDErrorMsg, resource.UserPrefix)
	}

	user, err := c.V2Client.GetIamUserById(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&userOut{
		Id:                   user.GetId(),
		Name:                 user.GetFullName(),
		Email:                user.GetEmail(),
		AuthenticationMethod: authMethodFormats[user.GetAuthType()],
	})
	return table.Print()
}
