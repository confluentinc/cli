package iam

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	invitationListFields    = []string{"Id", "Email", "FirstName", "LastName", "UserResourceId", "Status"}
	invitationHumanLabels   = []string{"ID", "Email", "First Name", "Last Name", "User ID", "Status"}
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

func newInvitationCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &invitationCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:         "invitation",
				Short:       "Manage invitations.",
				Args:        cobra.NoArgs,
			},
			prerunner,
		),
	}
	c.AddCommand(c.newInvitationListCommand())
	c.AddCommand(c.newInvitationCreateCommand())
	return c.Command
}

func (c invitationCommand) newInvitationListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the organization's invitations.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.listInvitations),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	return listCmd
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

func (c invitationCommand) newInvitationCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <email>",
		Short: "Invite a user to join your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.createInvitation),
	}
}

func (c invitationCommand) createInvitation(cmd *cobra.Command, args []string) error {
	email := args[0]
	matched := utils.ValidateEmail(email)
	if !matched {
		return errors.New(errors.BadEmailFormatErrorMsg)
	}
	newUser := &orgv1.User{Email: email}
	user, err := c.Client.User.CreateInvitation(context.Background(), &flowv1.CreateInvitationRequest{
		User: newUser,
		SendInvitation: true,
	})
	if err != nil {
		return err
	}
	utils.Println(cmd, fmt.Sprintf(errors.EmailInviteSentMsg, user.Email))
	return nil
}
