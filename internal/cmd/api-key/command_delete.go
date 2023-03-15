package apikey

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
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

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.ApiKey, apiKey)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}
	httpResp, err := c.V2Client.DeleteApiKey(apiKey)
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, deleteOperation, httpResp)
	}

	output.Printf(errors.DeletedResourceMsg, resource.ApiKey, apiKey)
	return c.keystore.DeleteAPIKey(apiKey)
}
