package iam

import (
	"context"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

var (
	humanLabelMap = map[string]string{
		"Id":                   "ID",
		"Email":                "Email",
		"FirstName":            "First Name",
		"LastName":             "Last Name",
		"Status":               "Status",
		"AuthenticationMethod": "Authentication Method",
	}
	structuredLabelMap = map[string]string{
		"Id":                   "id",
		"Email":                "email",
		"FirstName":            "first_name",
		"LastName":             "last_name",
		"Status":               "status",
		"AuthenticationMethod": "authentication_method",
	}
)

func (c userCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a user.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c userCommand) describe(cmd *cobra.Command, args []string) error {
	if resource.LookupType(args[0]) != resource.User {
		return errors.New(errors.BadResourceIDErrorMsg)
	}

	userProfile, err := c.Client.User.GetUserProfile(context.Background(), &orgv1.User{ResourceId: args[0]})
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
