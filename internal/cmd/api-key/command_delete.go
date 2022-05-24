package apikey

import (
	"context"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <api-key>",
		Short:             "Delete an API key.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]

	key, _, err := c.V2Client.GetApiKey(apiKey)
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, getOperation)
	}

	if isSchemaRegistryOrKsqlApiKey(key) {
		err = c.deleteV1(apiKey)
	} else {
		_, err = c.V2Client.DeleteApiKey(apiKey)
	}
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, deleteOperation)
	}

	utils.Printf(cmd, errors.DeletedAPIKeyMsg, apiKey)
	return c.keystore.DeleteAPIKey(apiKey)
}

func (c *command) deleteV1(apiKey string) error {
	userKey, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: apiKey, AccountId: c.EnvironmentId()})
	if err != nil {
		return err
	}
	key := &schedv1.ApiKey{
		Id:             userKey.Id,
		Key:            apiKey,
		AccountId:      c.EnvironmentId(),
		UserId:         userKey.UserId,
		UserResourceId: userKey.UserResourceId,
	}

	return c.Client.APIKey.Delete(context.Background(), key)
}
