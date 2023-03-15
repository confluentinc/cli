package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *userCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete users from your organization.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *userCommand) delete(cmd *cobra.Command, args []string) error {
	var errs error
	for _, resourceId := range args {
		if resource.LookupType(resourceId) != resource.User {
			errs = errors.Join(errs, errors.Errorf(errors.BadResourceIDErrorMsg, resource.UserPrefix))
		}
	}
	if errs != nil {
		return errs
	}

	fullName, validArgs, err := c.validateArgs(cmd, args)
	if err != nil {
		return err
	}
	args = validArgs

	if _, err := form.ConfirmDeletionType(cmd, resource.User, fullName, args); err != nil {
		return err
	}

	errs = nil
	var successful []string
	for _, resourceId := range args {
		if err := c.V2Client.DeleteIamUser(resourceId); err != nil {
			errs = errors.Join(errs, errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, resourceId, err))
		} else {
			successful = append(successful, resourceId)
		}
	}

	if len(successful) == 1 {
		output.Printf(errors.DeletedResourceMsg, resource.User, successful[0])
	} else if len(successful) > 1 {
		output.Printf(errors.DeletedResourcesMsg, resource.Plural(resource.User), utils.ArrayToCommaDelimitedString(successful, "and"))
	}

	return errs
}

func (c *userCommand) validateArgs(cmd *cobra.Command, args []string) (string, []string, error) {
	var fullName string
	describeFunc := func(arg string) error {
		if user, err := c.V2Client.GetIamUserById(arg); err != nil {
			return err
		} else if arg == args[0] {
			fullName = user.GetFullName()
		}
		return nil
	}

	validArgs, err := utils.ValidateArgsForDeletion(cmd, args, resource.User, describeFunc)

	return fullName, validArgs, err
}
