package apikey

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <api-key>",
		Short:             "Set the active API key for use in other commands.",
		Long:              "Set the active API key for use in any command which supports passing an API key with the `--api-key` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	cmd.Flags().String(resourceFlagName, "", "The resource ID.")
	cobra.CheckErr(cmd.MarkFlagRequired(resourceFlagName))

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()
	apiKey := args[0]
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.V2Client)
	if err != nil {
		return err
	}
	if resourceType != resource.KafkaCluster {
		return errors.Errorf(errors.NonKafkaNotImplementedErrorMsg)
	}
	cluster, err := c.Context.FindKafkaCluster(clusterId)
	if err != nil {
		return err
	}
	err = c.Context.UseAPIKey(apiKey, cluster.ID)
	if err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errors.APIKeyUseFailedErrorMsg, fmt.Sprintf(errors.APIKeyUseFailedSuggestions, apiKey))
	}
	output.Printf(errors.UseAPIKeyMsg, apiKey, clusterId)
	return nil
}
