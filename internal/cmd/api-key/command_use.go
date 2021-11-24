package apikey

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <api-key>",
		Short:             "Set the active API key for use in other commands.",
		Long:              "Set the active API key for use in any command which supports passing an API key with the `--api-key` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.use),
	}

	cmd.Flags().String(resourceFlagName, "", "The resource ID.")
	_ = cmd.MarkFlagRequired(resourceFlagName)

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	apiKey := args[0]

	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err != nil {
		return err
	}
	if resourceType != pcmd.KafkaResourceType {
		return errors.Errorf(errors.NonKafkaNotImplementedErrorMsg)
	}

	cluster, err := c.Context.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return err
	}

	if err := c.Context.UseAPIKey(cmd, apiKey, cluster.ID); err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errors.APIKeyUseFailedErrorMsg, fmt.Sprintf(errors.APIKeyUseFailedSuggestions, apiKey))
	}

	utils.Printf(cmd, errors.UseAPIKeyMsg, apiKey, clusterId)
	return nil
}
