package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listFields    = []string{"Id", "ResourceId", "Email", "FirstName", "LastName", "Status"}
	humanLabels   = []string{"Id", "Resource ID", "Email", "First Name", "Last Name", "Status"}
	humanLabelMap = map[string]string{
		"Id":         "Id",
		"ResourceId": "Resource ID",
		"Email":      "Email",
		"FirstName":  "First Name",
		"LastName":   "Last Name",
		"Status":     "Status",
	}
	structuredLabels   = []string{"id", "resource_id", "email", "first_name", "last_name", "status"}
	structuredLabelMap = map[string]string{
		"Id":         "id",
		"ResourceId": "resource_id",
		"Email":      "email",
		"FirstName":  "first_name",
		"LastName":   "last_name",
		"Status":     "status",
	}
	statusMap = map[flowv1.UserStatus]string{
		flowv1.UserStatus_USER_STATUS_UNKNOWN:     "Unknown",
		flowv1.UserStatus_USER_STATUS_UNVERIFIED:  "Unverified",
		flowv1.UserStatus_USER_STATUS_ACTIVE:      "Active",
		flowv1.UserStatus_USER_STATUS_DEACTIVATED: "Deactivated",
	}
)

type userCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type userStruct struct {
	Id         int32
	ResourceId string
	Email      string
	FirstName  string
	LastName   string
	Status     string
}

func NewUsersCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &userCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "user",
				Short: "Manage users.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}
	c.AddCommand(c.newUserDescribeCommand())
	c.AddCommand(c.newUserListCommand())
	c.AddCommand(c.newUserInviteCommand())
	c.AddCommand(c.newUserDeleteCommand())
	return c.Command
}

func (c userCommand) newUserDescribeCommand() *cobra.Command {
	describeCmd := &cobra.Command{
		Use:   "describe <resource id>",
		Short: "Describe a user.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	return describeCmd
}

func (c userCommand) describe(cmd *cobra.Command, args []string) error {
	resourceId := args[0]
	validFormat := strings.HasPrefix(resourceId, "u-")
	if !validFormat {
		return errors.New(errors.BadResourceIDErrorMsg)
	}
	user, err := c.Client.User.GetUserProfile(context.Background(), &orgv1.User{
		ResourceId: resourceId,
	})
	if err != nil {
		return err
	}

	// Avoid panics if new types of statuses are added in the future
	userStatus := "Unknown"
	if val, ok := statusMap[user.UserStatus]; ok {
		userStatus = val
	}

	return output.DescribeObject(cmd, &userStruct{
		ResourceId: user.ResourceId,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Status:     userStatus,
	}, listFields, humanLabelMap, structuredLabelMap)
}

func (c userCommand) newUserListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List an organization's users.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	return listCmd
}

func (c userCommand) list(cmd *cobra.Command, _ []string) error {
	users, err := c.Client.User.List(context.Background())
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}
	for _, user := range users {
		userProfile, err := c.Client.User.GetUserProfile(context.Background(), &orgv1.User{
			ResourceId: user.ResourceId,
		})
		if err != nil {
			return err
		}

		// Avoid panics if new types of statuses are added in the future
		userStatus := "Unknown"
		if val, ok := statusMap[userProfile.UserStatus]; ok {
			userStatus = val
		}

		outputWriter.AddElement(&userStruct{
			Id:         user.Id,
			ResourceId: userProfile.ResourceId,
			Email:      userProfile.Email,
			FirstName:  userProfile.FirstName,
			LastName:   userProfile.LastName,
			Status:     userStatus,
		})
	}
	return outputWriter.Out()
}

func (c userCommand) newUserInviteCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "invite <email>",
		Short: "Invite a user to join your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.invite),
	}
	return createCmd
}

func (c userCommand) invite(cmd *cobra.Command, args []string) error {
	email := args[0]
	matched := utils.ValidateEmail(email)
	if !matched {
		return errors.New(errors.BadEmailFormatErrorMsg)
	}
	newUser := &orgv1.User{Email: email, OrganizationId: c.State.Auth.User.OrganizationId}
	user, err := c.Client.User.Invite(context.Background(), newUser)
	if err != nil {
		return err
	}
	utils.Println(cmd, fmt.Sprintf(errors.EmailInviteSentMsg, user.Email))
	return nil
}

func (c userCommand) newUserDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <resource id>",
		Short: "Delete a user from your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	return deleteCmd
}

func (c userCommand) delete(cmd *cobra.Command, args []string) error {
	resourceId := args[0]
	validFormat := strings.HasPrefix(resourceId, "u-")
	if !validFormat {
		return errors.New(errors.BadResourceIDErrorMsg)
	}
	err := c.Client.User.Delete(context.Background(), &orgv1.User{
		ResourceId:     resourceId,
		OrganizationId: c.State.Auth.User.OrganizationId,
	})
	if err != nil {
		return err
	}
	utils.Println(cmd, fmt.Sprintf(errors.DeletedUserMsg, resourceId))
	return nil
}
