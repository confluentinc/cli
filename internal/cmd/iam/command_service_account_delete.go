package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) delete(cmd *cobra.Command, args []string) error {
	displayName, validArgs, err := c.validateArgs(cmd, args)
	if err != nil {
		return err
	}
	args = validArgs

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.ServiceAccount, args[0], displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ServiceAccount, args); err != nil || !ok {
			return err
		}
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteIamServiceAccount(id); err != nil {
			errs = errors.Join(errs, errors.Errorf(errors.DeleteResourceErrorMsg, resource.ServiceAccount, id, err))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ServiceAccount)

	return errs
}

func (c *serviceAccountCommand) validateArgs(cmd *cobra.Command, args []string) (string, []string, error) {
	if err := resource.ValidatePrefixes(resource.ServiceAccount, args); err != nil {
		return "", nil, err
	}

	var displayName string
	describeFunc := func(id string) error {
		if serviceAccount, _, err := c.V2Client.GetIamServiceAccount(id); err != nil {
			return err
		} else if displayName == "" { // store the first valid provider name
			displayName = serviceAccount.GetDisplayName()
		}
		return nil
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.ServiceAccount, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.ServiceAccount, "iam service-account"))

	return displayName, validArgs, err
}
