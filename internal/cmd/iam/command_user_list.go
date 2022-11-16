package iam

import (
	"context"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
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

	list := output.NewList(cmd)
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
		for _, method := range userProfile.GetAuthConfig().GetAllowedAuthMethods() {
			authMethods = append(authMethods, authMethodFormats[method])
		}

		list.Add(&userOut{
			Id:                   userProfile.ResourceId,
			Name:                 getName(userProfile),
			Email:                userProfile.Email,
			Status:               userStatus,
			AuthenticationMethod: strings.Join(authMethods, ", "),
		})
	}
	return list.Print()
}
