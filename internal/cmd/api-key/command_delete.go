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
		Use:               "delete <api-key-1> [api-key-2] ... [api-key-n]",
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

	if validArgs, err := c.validateArgs(cmd, args); err != nil {
		return err
	} else {
		args = validArgs
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

func (c *command) validateArgs(cmd *cobra.Command, args []string) ([]string, error) {
	// Single
	if len(args) == 1 {
		if _, _, err := c.V2Client.GetApiKey(args[0]); err != nil {
			return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.ApiKey, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.ApiKey))
		}
		return args, nil
	}

	// Multiple
	setFunc := func() (types.Set, error) {
		apiKeys, err := c.V2Client.ListApiKeys("", "")
		if err != nil {
			return nil, err
		}

		set := types.NewSet()
		for _, apiKey := range apiKeys {
			set.Add(apiKey.GetId())
		}

		return set, nil
	}

	return utils.ValidateArgsForDeletion(cmd, args, resource.ApiKey, setFunc)
}
