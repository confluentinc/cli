package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *serviceAccountCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete service accounts.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete service account "sa-123456".`,
				Code: "confluent iam service-account delete sa-123456",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) delete(cmd *cobra.Command, args []string) error {
	var errs error
	for _, serviceAccountId := range args {
		if resource.LookupType(serviceAccountId) != resource.ServiceAccount {
			errs = errors.Join(errs, errors.New(errors.BadServiceAccountIDErrorMsg))
		}
	}
	if errs != nil {
		return errs
	}

	displayName, err := c.checkExistence(cmd, args)
	if err != nil {
		return err
	}

	if _, err := form.ConfirmDeletionType(cmd, resource.ServiceAccount, displayName, args); err != nil {
		return err
	}

	errs = nil
	for _, serviceAccountId := range args {
		if err := c.V2Client.DeleteIamServiceAccount(serviceAccountId); err != nil {
			errs = errors.Join(errs, errors.Errorf(errors.DeleteResourceErrorMsg, resource.ServiceAccount, serviceAccountId, err))
		} else {
			output.ErrPrintf(errors.DeletedResourceMsg, resource.ServiceAccount, serviceAccountId)
		}
	}

	return errs
}

func (c *serviceAccountCommand) checkExistence(cmd *cobra.Command, args []string) (string, error) {
	// Single
	if len(args) == 1 {
		if serviceAccount, _, err := c.V2Client.GetIamServiceAccount(args[0]); err != nil {
			return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.ServiceAccount, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ServiceAccount))
		} else {
			return serviceAccount.GetDisplayName(), nil
		}
	}

	// Multiple
	serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
	if err != nil {
		return "", err
	}

	set := types.NewSet()
	for _, serviceAccount := range serviceAccounts {
		set.Add(serviceAccount.GetId())
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return "", err
	} else if force && len(invalidArgs) > 0 {
		args = validArgs
		return "", nil
	}

	invalidArgsStr := utils.ArrayToCommaDelimitedString(invalidArgs, "and")
	if len(invalidArgs) == 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.ServiceAccount, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ServiceAccount))
	} else if len(invalidArgs) > 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resource.ServiceAccount), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ServiceAccount))
	}

	return "", nil
}
