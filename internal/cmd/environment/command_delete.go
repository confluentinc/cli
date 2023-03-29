package environment

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete Confluent Cloud environments and all of their resources.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	displayName, validArgs, err := c.validateArgs(cmd, args)
	if err != nil {
		return err
	}
	args = validArgs

	if _, err := form.ConfirmDeletionType(cmd, resource.Environment, displayName, args); err != nil {
		return err
	}

	var errs error
	var deleted []string
	environmentId, _ := c.EnvironmentId()
	for _, id := range args {
		if httpResp, err := c.V2Client.DeleteOrgEnvironment(id); err != nil {
			errs = errors.Join(errs, errors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp))
		} else {
			deleted = append(deleted, id)
			if id == environmentId {
				c.Context.SetEnvironment(nil)

				if err := c.Config.Save(); err != nil {
					errs = errors.Join(errs, errors.Wrap(err, errors.EnvSwitchErrorMsg))
				}
			}
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.Environment)

	if errs != nil {
		errs = errors.NewErrorWithSuggestions(errs.Error(), fmt.Sprintf(errors.ListResourceSuggestions, resource.Environment, resource.Environment))
	}

	return errs
}

func (c *command) validateArgs(cmd *cobra.Command, args []string) (string, []string, error) {
	var displayName string
	describeFunc := func(id string) error {
		if environment, _, err := c.V2Client.GetOrgEnvironment(id); err != nil {
			return err
		} else if id == args[0] {
			displayName = environment.GetDisplayName()
		}
		return nil
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.Environment, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.Environment, resource.Environment))

	return displayName, validArgs, err
}
