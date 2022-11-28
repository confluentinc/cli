package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c userCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a user from your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c userCommand) delete(cmd *cobra.Command, args []string) error {
	resourceId := args[0]
	if resource.LookupType(resourceId) != resource.User {
		return fmt.Errorf(errors.BadResourceIDErrorMsg, resource.UserPrefix)
	}

	user, err := c.V2Client.GetIamUserById(resourceId)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.User, resourceId, user.GetFullName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, user.GetFullName()); err != nil {
		return err
	}

	err = c.V2Client.DeleteIamUser(resourceId)
	if err != nil {
		return errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, resourceId, err)
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.User, resourceId)
	return nil
}
