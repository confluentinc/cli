package iam

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type invitationOut struct {
	Id             string `human:"ID" structured:"id"`
	Email          string `human:"Email" structured:"email"`
	FirstName      string `human:"First Name" structured:"first_name"`
	LastName       string `human:"Last Name" structured:"last_name"`
	UserResourceId string `human:"User ID" structured:"user_resource_id"`
	Status         string `human:"Status" structured:"status"`
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
	invitations, err := c.Client.User.ListInvitations(context.Background())
	if err != nil {
		return err
	}

	if len(invitations) == 0 {
		utils.Println(cmd, "No invitations found.")
		return nil
	}

	list := output.NewList(cmd)
	for _, invitation := range invitations {
		user := &orgv1.User{ResourceId: invitation.UserResourceId}

		var firstName, lastName string
		if user, err = c.Client.User.Describe(context.Background(), user); err == nil {
			firstName = user.FirstName
			lastName = user.LastName
		}

		list.Add(&invitationOut{
			Id:             invitation.Id,
			Email:          invitation.Email,
			FirstName:      firstName,
			LastName:       lastName,
			UserResourceId: invitation.UserResourceId,
			Status:         invitation.Status,
		})
	}
	return list.Print()
}
