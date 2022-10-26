package iam

import (
	"context"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	humanLabels      = []string{"ID", "Email", "First Name", "Last Name", "Status", "Authentication Method"}
	structuredLabels = []string{"id", "email", "first_name", "last_name", "status", "authentication_method"}
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

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	for _, user := range users {
		userProfile, err := c.Client.User.GetUserProfile(context.Background(), &orgv1.User{ResourceId: *user.Id})
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
