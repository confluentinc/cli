package apikey

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <api-key>",
		Short:             "Delete an API key.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]

	key, httpResp, err := c.V2Client.GetApiKey(apiKey)
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, getOperation, httpResp)
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.ApiKey, apiKey)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	if isSchemaRegistryOrKsqlApiKey(key) {
		err = c.deleteV1(apiKey)
	} else {
		httpResp, err = c.V2Client.DeleteApiKey(apiKey)
	}
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, deleteOperation, httpResp)
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.ApiKey, apiKey)
	return c.keystore.DeleteAPIKey(apiKey)
}

func (c *command) deleteV1(apiKey string) error {
	userKey, err := c.PrivateClient.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: apiKey, AccountId: c.EnvironmentId()})
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

	return c.PrivateClient.APIKey.Delete(context.Background(), key)
}
