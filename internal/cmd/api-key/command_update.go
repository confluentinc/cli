package apikey

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <api-key>",
		Short: "Update an API key.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
	}

	cmd.Flags().String("description", "", "Description of the API key.")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]
	key, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: apiKey, AccountId: c.EnvironmentId()})
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("description") {
		key.Description = description
	}

	err = c.Client.APIKey.Update(context.Background(), key)
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("description") {
		utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "API key", apiKey, description)
	}
	return nil
}
