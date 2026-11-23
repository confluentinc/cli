package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type invitationOut struct {
	Id     string `human:"ID" serialized:"id"`
	Name   string `human:"Name" serialized:"name"`
	Email  string `human:"Email" serialized:"email"`
	UserId string `human:"User" serialized:"user_id"`
	Status string `human:"Status" serialized:"status"`
}

func (c invitationCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the organization's invitations.",
		Args:  cobra.NoArgs,
		RunE:  c.listInvitations,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c invitationCommand) listInvitations(cmd *cobra.Command, _ []string) error {
	invitations, err := c.V2Client.ListIamInvitations()
	if err != nil {
		return err
	}

	users, err := c.V2Client.ListIamUsers()
	if err != nil {
		return err
	}

	names := make(map[string]string, len(users))
	for _, user := range users {
		names[user.GetId()] = user.GetFullName()
	}

	list := output.NewList(cmd)
	for _, invitation := range invitations {
		// Invitations whose user is not found get an empty name
		name := names[invitation.User.GetId()]

		list.Add(&invitationOut{
			Id:     invitation.GetId(),
			Name:   name,
			Email:  invitation.GetEmail(),
			UserId: invitation.User.GetId(),
			Status: invitation.GetStatus(),
		})
	}
	return list.Print()
}
