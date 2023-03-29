package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *userCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete users from your organization.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *userCommand) delete(cmd *cobra.Command, args []string) error {
	fullName, validArgs, err := c.validateArgs(cmd, args)
	if err != nil {
		return err
	}
	args = validArgs

	if _, err := form.ConfirmDeletionType(cmd, resource.User, fullName, args); err != nil {
		return err
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteIamUser(id); err != nil {
			errs = errors.Join(errs, errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, id, err))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.User)

	return errs
}

func (c *userCommand) validateArgs(cmd *cobra.Command, args []string) (string, []string, error) {
	if err := resource.ValidatePrefixes(resource.User, args); err != nil {
		return "", nil, err
	}

	var fullName string
	describeFunc := func(id string) error {
		if user, err := c.V2Client.GetIamUserById(id); err != nil {
			return err
		} else if id == args[0] {
			fullName = user.GetFullName()
		}
		return nil
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.User, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.User, "iam user"))

	return fullName, validArgs, err
}
