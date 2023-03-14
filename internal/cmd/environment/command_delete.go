package environment

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

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete Confluent Cloud environments and all of their resources.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	displayName, err := c.checkExistence(cmd, args)
	if err != nil {
		return err
	}

	if _, err := form.ConfirmDeletionType(cmd, resource.Environment, displayName, args); err != nil {
		return err
	}

	var errs error
	for _, envId := range args {
		if httpResp, err := c.V2Client.DeleteOrgEnvironment(envId); err != nil {
			errs = errors.Join(errs, errors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp))
		} else {
			output.ErrPrintf(errors.DeletedResourceMsg, resource.Environment, envId)
			if envId == c.EnvironmentId() {
				c.Context.SetEnvironment(nil)

				if err := c.Config.Save(); err != nil {
					errs = errors.Join(errs, errors.Wrap(err, errors.EnvSwitchErrorMsg))
				}
			}
		}
	}
	if errs != nil {
		errs = errors.NewErrorWithSuggestions(errs.Error(), fmt.Sprintf(errors.OrgResourceNotFoundSuggestions, resource.Environment))
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) (string, error) {
	// Single
	if len(args) == 1 {
		if environment, _, err := c.V2Client.GetOrgEnvironment(args[0]); err != nil {
			return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Environment, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Environment))
		} else {
			return environment.GetDisplayName(), nil
		}
	}

	// Multiple
	environments, err := c.V2Client.ListOrgEnvironments()
	if err != nil {
		return "", err
	}

	set := types.NewSet()
	for _, environment := range environments {
		set.Add(environment.GetId())
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
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Environment, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Environment))
	} else if len(invalidArgs) > 1 {
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, utils.Plural(resource.Environment), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Environment))
	}

	return "", nil
}
