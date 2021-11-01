package iam

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
	listFields    = []string{"Id", "Email", "FirstName", "LastName", "Status", "AuthenticationMethod"}
	humanLabels   = []string{"ID", "Email", "First Name", "Last Name", "Status", "Authentication Method"}
	humanLabelMap = map[string]string{
		"Id":                   "ID",
		"Email":                "Email",
		"FirstName":            "First Name",
		"LastName":             "Last Name",
		"Status":               "Status",
		"AuthenticationMethod": "Authentication Method",
	}
	structuredLabels   = []string{"id", "email", "first_name", "last_name", "status", "authentication_method"}
	structuredLabelMap = map[string]string{
		"Id":                   "id",
		"Email":                "email",
		"FirstName":            "first_name",
		"LastName":             "last_name",
		"Status":               "status",
		"AuthenticationMethod": "authentication_method",
	}
	statusMap = map[flowv1.UserStatus]string{
		flowv1.UserStatus_USER_STATUS_UNKNOWN:     "Unknown",
		flowv1.UserStatus_USER_STATUS_UNVERIFIED:  "Unverified",
		flowv1.UserStatus_USER_STATUS_ACTIVE:      "Active",
		flowv1.UserStatus_USER_STATUS_DEACTIVATED: "Deactivated",
	}
)

var authMethodFormats = map[orgv1.AuthMethod]string{
	orgv1.AuthMethod_AUTH_METHOD_UNKNOWN:      "Unknown",
	orgv1.AuthMethod_AUTH_METHOD_USERNAME_PWD: "Username/Password",
	orgv1.AuthMethod_AUTH_METHOD_SSO:          "SSO",
}

type userCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type userStruct struct {
	Id                   string
	Email                string
	FirstName            string
	LastName             string
	Status               string
	AuthenticationMethod string
}

func NewUserCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &userCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:         "user",
				Short:       "Manage users.",
				Args:        cobra.NoArgs,
				Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
			},
			prerunner,
		),
	}
	c.AddCommand(c.newUserDescribeCommand())
	c.AddCommand(c.newUserListCommand())
	c.AddCommand(c.newUserDeleteCommand())
	c.AddCommand(NewInvitationCommand(prerunner))
	return c.Command
}

func (c userCommand) newUserDescribeCommand() *cobra.Command {
	describeCmd := &cobra.Command{
		Use:   "describe <id>",
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
	userProfile, err := c.Client.User.GetUserProfile(context.Background(), &orgv1.User{
		ResourceId: resourceId,
	})
	if err != nil {
		return err
	}

	// Avoid panics if new types of statuses are added in the future
	userStatus := "Unknown"
	if val, ok := statusMap[userProfile.UserStatus]; ok {
		userStatus = val
	}

	var authMethods []string
	if userProfile.GetAuthConfig() != nil {
		for _, method := range userProfile.GetAuthConfig().AllowedAuthMethods {
			authMethods = append(authMethods, authMethodFormats[method])
		}
	}

	return output.DescribeObject(cmd, &userStruct{
		Id:                   userProfile.ResourceId,
		Email:                userProfile.Email,
		FirstName:            userProfile.FirstName,
		LastName:             userProfile.LastName,
		Status:               userStatus,
		AuthenticationMethod: strings.Join(authMethods, ", "),
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

		var authMethods []string
		if userProfile.GetAuthConfig() != nil {
			for _, method := range userProfile.GetAuthConfig().AllowedAuthMethods {
				authMethods = append(authMethods, authMethodFormats[method])
			}
		}

		outputWriter.AddElement(&userStruct{
			Id:                   userProfile.ResourceId,
			Email:                userProfile.Email,
			FirstName:            userProfile.FirstName,
			LastName:             userProfile.LastName,
			Status:               userStatus,
			AuthenticationMethod: strings.Join(authMethods, ", "),
		})
	}
	return outputWriter.Out()
}

func (c userCommand) newUserDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
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
		ResourceId: resourceId,
	})
	if err != nil {
		return err
	}
	utils.Println(cmd, fmt.Sprintf(errors.DeletedUserMsg, resourceId))
	return nil
}
