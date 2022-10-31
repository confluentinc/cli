package iam

import (
	"context"
	"fmt"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c userCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a user.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c userCommand) describe(cmd *cobra.Command, args []string) error {
	if resource.LookupType(args[0]) != resource.User {
		return fmt.Errorf(errors.BadResourceIDErrorMsg, resource.UserPrefix)
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
	for _, method := range userProfile.GetAuthConfig().GetAllowedAuthMethods() {
		authMethods = append(authMethods, authMethodFormats[method])
	}

	table := output.NewTable(cmd)
	table.Add(&userOut{
		Id:                   userProfile.ResourceId,
		Email:                userProfile.Email,
		FirstName:            userProfile.FirstName,
		LastName:             userProfile.LastName,
		Status:               userStatus,
		AuthenticationMethod: strings.Join(authMethods, ", "),
	})
	return table.Print()
}
