package apikey

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <api-key>",
		Short: "Delete an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]

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

	err = c.Client.APIKey.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.DeletedAPIKeyMsg, apiKey)
	err = c.keystore.DeleteAPIKey(apiKey)
	if err != nil {
		return err
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, key.UserResourceId)
	c.analyticsClient.SetSpecialProperty(analytics.ApiKeyPropertiesKey, key.Key)
	return nil
}
