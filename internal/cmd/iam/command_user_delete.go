package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
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

	fullName, err := c.checkExistence(cmd, args)
	if err != nil {
		return err
	}

	if _, err := form.ConfirmDeletionType(cmd, resource.User, fullName, args); err != nil {
		return err
	}

	errs = nil
	for _, resourceId := range args {
		if err := c.V2Client.DeleteIamUser(resourceId); err != nil {
			errs = errors.Join(errs, errors.Errorf(errors.DeleteResourceErrorMsg, resource.User, resourceId, err))
		} else {
			output.Printf(errors.DeletedResourceMsg, resource.User, resourceId)
		}
	}

	return errs
}

func (c *userCommand) checkExistence(cmd *cobra.Command, args []string) (string, error) {
	// Single
	if len(args) == 1 {
		if user, err := c.V2Client.GetIamUserById(args[0]); err != nil {
			return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.User, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.User))
		} else {
			return user.GetFullName(), nil
		}
	}

	// Multiple
	users, err := c.V2Client.ListIamUsers()
	if err != nil {
		return "", err
	}

	set := types.NewSet()
	for _, user := range users {
		set.Add(user.GetId())
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return "", err
	} else if force && len(invalidArgs) > 0 {
		args = validArgs
		return "", nil
	}

	invalidArgsStr := utils.ArrayToCommaDelimitedStringWithAnd(invalidArgs)
	if len(invalidArgs) == 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.User, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.User))
	} else if len(invalidArgs) > 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, utils.Plural(resource.User), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.User))
	}

	return "", nil
}
