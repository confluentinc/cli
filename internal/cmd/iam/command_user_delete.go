package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *userCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete users from your organization.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *userCommand) delete(cmd *cobra.Command, args []string) error {
	fullName, err := c.validateArgs(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.User, args[0], fullName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.User, args); err != nil || !ok {
			return err
		}
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

func (c *userCommand) validateArgs(cmd *cobra.Command, args []string) (string, error) {
	if err := resource.ValidatePrefixes(resource.User, args); err != nil {
		return "", err
	}

	var fullName string
	describeFunc := func(id string) error {
		user, err := c.V2Client.GetIamUserById(id)
		if err == nil && fullName == "" { // store the first valid user name
			fullName = user.GetFullName()
		}
		return err
	}

	return fullName, deletion.ValidateArgsForDeletion(cmd, args, resource.User, describeFunc)
}
