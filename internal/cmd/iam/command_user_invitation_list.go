package iam

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	invitationListFields       = []string{"Id", "Email", "FirstName", "LastName", "UserResourceId", "Status"}
	invitationHumanLabels      = []string{"ID", "Email", "First Name", "Last Name", "User ID", "Status"}
	invitationStructuredLabels = []string{"id", "email", "first_name", "last_name", "user_resource_id", "status"}
)

type invitationStruct struct {
	Id             string
	Email          string
	FirstName      string
	LastName       string
	UserResourceId string
	Status         string
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

	outputWriter, err := output.NewListOutputWriter(cmd, invitationListFields, invitationHumanLabels, invitationStructuredLabels)
	if err != nil {
		return err
	}

	for _, invitation := range invitations {
		user := &orgv1.User{ResourceId: invitation.User.GetId()}

		var firstName, lastName string
		if user, err = c.PrivateClient.User.Describe(context.Background(), user); err == nil {
			firstName = user.FirstName
			lastName = user.LastName
		}

		outputWriter.AddElement(&invitationStruct{
			Id:             invitation.GetId(),
			Email:          invitation.GetEmail(),
			FirstName:      firstName,
			LastName:       lastName,
			UserResourceId: invitation.User.GetId(),
			Status:         invitation.GetStatus(),
		})
	}

	return outputWriter.Out()
}
