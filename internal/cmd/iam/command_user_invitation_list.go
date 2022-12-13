package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
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

	if len(invitations) == 0 {
		utils.Println(cmd, "No invitations found.")
		return nil
	}

	list := output.NewList(cmd)
	for _, invitation := range invitations {
		var name string
		if user, err := c.V2Client.GetIamUserById(invitation.User.GetId()); err == nil {
			name = user.GetFullName()
		}

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
