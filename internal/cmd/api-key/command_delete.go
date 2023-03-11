package apikey

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	perrors "github.com/confluentinc/cli/internal/pkg/errors"
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
			errs = errors.Join(errs, perrors.CatchApiKeyForbiddenAccessError(err, deleteOperation, id, httpResp))
		} else {
			output.Printf(perrors.DeletedResourceMsg, resource.ApiKey, id)
			if err := c.keystore.DeleteAPIKey(id); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	if errs != nil {
		errs = perrors.NewErrorWithSuggestions(errs.Error(), perrors.APIKeyNotFoundSuggestions)
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) error {
	// Single
	if len(args) == 1 {
		if _, httpResp, err := c.V2Client.GetApiKey(args[0]); err != nil {
			return perrors.CatchApiKeyForbiddenAccessError(err, getOperation, args[0], httpResp)
		}
		return nil
	}

	// Multiple
	apiKeys, err := c.V2Client.ListApiKeys("", "")
	if err != nil {
		return err
	}

	apiKeySet := types.NewSet()
	for _, apiKey := range apiKeys {
		apiKeySet.Add(apiKey.GetId())
	}

	invalidKeys := apiKeySet.Difference(args)
	if len(invalidKeys) > 0 {
		return perrors.NewErrorWithSuggestions(fmt.Sprintf(perrors.AccessForbiddenErrorMsg, resource.ApiKey, utils.ArrayToCommaDelimitedStringWithAnd(invalidKeys)), perrors.APIKeyNotFoundSuggestions)
	}

	return nil
}
