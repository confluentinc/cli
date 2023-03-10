package environment

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	perrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-N]",
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

	promptMsg := fmt.Sprintf(perrors.DeleteResourceConfirmMsg, resource.Environment, args[0], displayName)
	if _, err := form.ConfirmDeletionTemp(cmd, promptMsg, displayName, resource.Environment, args); err != nil {
		return err
	}

	var errs error
	for _, envId := range args {
		if httpResp, err := c.V2Client.DeleteOrgEnvironment(envId); err != nil {
			errs = errors.Join(errs, perrors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp))
		} else {
			output.ErrPrintf(perrors.DeletedResourceMsg, resource.Environment, envId)
			if envId == c.EnvironmentId() {
				c.Context.SetEnvironment(nil)

				if err := c.Config.Save(); err != nil {
					errs = errors.Join(errs, perrors.Wrap(err, perrors.EnvSwitchErrorMsg))
				}
			}
		}
	}
	if errs != nil {
		errs = perrors.NewErrorWithSuggestions(errs.Error(), fmt.Sprintf(perrors.OrgResourceNotFoundSuggestions, resource.Environment))
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) (string, error) {
	// Single
	if len(args) == 1 {
		if environment, httpResp, err := c.V2Client.GetOrgEnvironment(args[0]); err != nil {
			return "", perrors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp)
		} else {
			return environment.GetDisplayName(), nil
		}
	}

	// Multiple
	environments, err := c.V2Client.ListOrgEnvironments()
	if err != nil {
		return "", err
	}

	environmentSet := set.New()
	for _, environment := range environments {
		environmentSet.Add(environment.GetId())
	}

	invalidEnvironments := environmentSet.Difference(args)
	if len(invalidEnvironments) > 0 {
		return "", perrors.NewErrorWithSuggestions("environment(s) not found or access forbidden: " + utils.ArrayToCommaDelimitedStringWithAnd(invalidEnvironments), fmt.Sprintf(perrors.OrgResourceNotFoundSuggestions, resource.Environment))
	}

	return "", nil
}
