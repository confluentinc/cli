package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c userCommand) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a user from your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}
}

func (c userCommand) delete(cmd *cobra.Command, args []string) error {
	resourceId := args[0]
	if resource.LookupType(resourceId) != resource.User {
		return errors.New(errors.BadResourceIDErrorMsg)
	}

	_, err := c.V2Client.DeleteIamUser(resourceId)
	if err != nil {
		return errors.Errorf(`failed to delete user "%s": %v`, resourceId, err)
	}

	utils.Println(cmd, fmt.Sprintf(errors.DeletedUserMsg, resourceId))
	return nil
}
