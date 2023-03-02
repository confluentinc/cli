package apikey

import (
	"github.com/spf13/cobra"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
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

	if cmd.Flags().Changed("description") {
		apiKeyUpdate := apikeysv2.IamV2ApiKeyUpdate{
			Spec: &apikeysv2.IamV2ApiKeySpecUpdate{Description: apikeysv2.PtrString(description)},
		}
		_, httpResp, err := c.V2Client.UpdateApiKey(apiKey, apiKeyUpdate)

		if err != nil {
			return errors.CatchApiKeyForbiddenAccessError(err, updateOperation, httpResp)
		}

		output.ErrPrintf(errors.UpdateSuccessMsg, "description", "API key", apiKey, description)
	}

	return nil
}
