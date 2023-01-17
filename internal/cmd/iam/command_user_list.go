package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c userCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List an organization's users.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c userCommand) list(cmd *cobra.Command, _ []string) error {
	users, err := c.V2Client.ListIamUsers()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, user := range users {
		list.Add(&userOut{
			Id:                   user.GetId(),
			Name:                 user.GetFullName(),
			Email:                user.GetEmail(),
			AuthenticationMethod: authMethodFormats[user.GetAuthType()],
		})
	}
	return list.Print()
}
