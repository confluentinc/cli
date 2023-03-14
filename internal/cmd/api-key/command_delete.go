package apikey

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
		Use:               "delete <api-key-1> [api-key-2] ... [api-key-N]",
		Short:             "Delete API keys.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	if err := c.checkExistence(cmd, args); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ApiKey, args); err != nil || !ok {
		return err
	}

	var errs error
	for _, id := range args {
		if httpResp, err := c.V2Client.DeleteApiKey(id); err != nil {
			errs = errors.Join(errs, errors.CatchApiKeyForbiddenAccessError(err, deleteOperation, id, httpResp))
		} else {
			output.Printf(errors.DeletedResourceMsg, resource.ApiKey, id)
			if err := c.keystore.DeleteAPIKey(id); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	if errs != nil {
		errs = errors.NewErrorWithSuggestions(errs.Error(), errors.APIKeyNotFoundSuggestions)
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) error {
	// Single
	if len(args) == 1 {
		if _, _, err := c.V2Client.GetApiKey(args[0]); err != nil {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.ApiKey, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ApiKey))
		}
		return nil
	}

	// Multiple
	apiKeys, err := c.V2Client.ListApiKeys("", "")
	if err != nil {
		return err
	}

	set := types.NewSet()
	for _, apiKey := range apiKeys {
		set.Add(apiKey.GetId())
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return err
	} else if force && len(invalidArgs) > 0 {
		args = validArgs
		return nil
	}

	invalidArgsStr := utils.ArrayToCommaDelimitedStringWithAnd(invalidArgs)
	if len(invalidArgs) == 1 {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.ApiKey, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ApiKey))
	} else if len(invalidArgs) > 1 {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, utils.Plural(resource.ApiKey), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ApiKey))
	}

	return nil
}
