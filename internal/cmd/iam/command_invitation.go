package iam

import (
	"context"
	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	invitationListFields    = []string{"Id", "Email", "FirstName", "LastName", "UserResourceId", "Status"}
	invitationHumanLabels   = []string{"ID", "Email", "First Name", "Last Name", "User Resource ID", "Status"}
	invitationStructuredLabels   = []string{"id", "email", "first_name", "last_name", "user_resource_id", "status"}
)

type invitationCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type invitationStruct struct {
	Id                   string
	Email                string
	FirstName            string
	LastName             string
	UserResourceId       string
	Status               string
}

func NewInvitationCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &invitationCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:         "invitation",
				Short:       "Manage invitation.",
				Args:        cobra.NoArgs,
				Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
			},
			prerunner,
		),
	}
	c.AddCommand(c.newInvitationListCommand())
	return c.Command
}

func (c invitationCommand) newInvitationListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the organization's invitations.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.invitationList),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	return listCmd
}

func (c invitationCommand) invitationList(cmd *cobra.Command, _ []string) error {
	invitations, err := c.Client.User.ListInvitations(context.Background())
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
		userProfile, err := c.Client.User.GetUserProfile(context.Background(), &orgv1.User{
			ResourceId: invitation.UserResourceId,
		})
		if err != nil {
			return err
		}
		outputWriter.AddElement(&invitationStruct{
			Id:             invitation.Id,
			Email:          invitation.Email,
			FirstName:      userProfile.FirstName,
			LastName:       userProfile.LastName,
			UserResourceId: invitation.UserResourceId,
			Status:         invitation.Status,
		})
	}
	return outputWriter.Out()
}
