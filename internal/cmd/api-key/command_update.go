package apikey

import (
	"context"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <api-key>",
		Short:             "Update an API key.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}

	cmd.Flags().String("description", "", "Description of the API key.")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	key, httpResp, err := c.V2Client.GetApiKey(apiKey)
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, httpResp)
	}

	if cmd.Flags().Changed("description") {
		if isSchemaRegistryOrKsqlApiKey(key) {
			err = c.v1Update(apiKey, description)
		} else {
			apiKeyUpdate := apikeysv2.IamV2ApiKeyUpdate{
				Spec: &apikeysv2.IamV2ApiKeySpecUpdate{Description: apikeysv2.PtrString(description)},
			}
			_, httpResp, err = c.V2Client.UpdateApiKey(apiKey, apiKeyUpdate)
		}
		if err != nil {
			return errors.CatchApiKeyForbiddenAccessError(err, httpResp)
		}

		utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "API key", apiKey, description)
	}

	return nil
}

func (c *command) v1Update(apiKey, description string) error {
	key, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: apiKey, AccountId: c.EnvironmentId()})
	if err != nil {
		return err
	}

	key.Description = description
	return c.Client.APIKey.Update(context.Background(), key)
}
