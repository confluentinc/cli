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
		Short:             "Use an API key in subsequent commands.",
		Long:              "Choose an API key to be used in subsequent commands which support passing an API key with the `--api-key` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	// Deprecated
	c.addResourceFlag(cmd)
	cobra.CheckErr(cmd.Flags().MarkHidden("resource"))

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	var clusterId string

	if cmd.Flags().Changed("resource") {
		resourceId, _, err := c.resolveResourceId(cmd, c.V2Client)
		if err != nil {
			return err
		}
		if resource.LookupType(resourceId) != resource.KafkaCluster {
			return errors.Errorf(errors.NonKafkaNotImplementedErrorMsg)
		}
		clusterId = resourceId
	} else {
		clusterId = c.Context.KafkaClusterContext.FindApiKeyClusterId(args[0])
		if clusterId == "" {
			return fmt.Errorf(`could not find resource for API key "%s"`, args[0])
		}
	}

	if err := c.Context.UseAPIKey(args[0], clusterId); err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errors.APIKeyUseFailedErrorMsg, fmt.Sprintf(errors.APIKeyUseFailedSuggestions, args[0]))
	}

	output.Printf(errors.UseAPIKeyMsg, args[0])
	return nil
}
