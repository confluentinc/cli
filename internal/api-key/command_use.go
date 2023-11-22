package apikey

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

const useAPIKeyMsg = "Using API Key \"%s\".\n"

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
	c.addResourceFlag(cmd, true)
	cobra.CheckErr(cmd.Flags().MarkHidden("resource"))

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	var clusterId string

	if cmd.Flags().Changed("resource") {
		_, resourceId, _, err := c.resolveResourceId(cmd, c.V2Client)
		if err != nil {
			return err
		}
		if resource.LookupType(resourceId) != resource.KafkaCluster {
			return fmt.Errorf(nonKafkaNotImplementedErrorMsg)
		}
		clusterId = resourceId
	} else {
		clusterId = c.Context.KafkaClusterContext.FindApiKeyClusterId(args[0])
		if clusterId == "" {
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(`API key "%s" and associated Kafka cluster are not stored in local CLI state`, args[0]),
				fmt.Sprintf(apiKeyUseFailedSuggestions, args[0]),
			)
		}
	}

	if err := c.useAPIKey(args[0], clusterId); err != nil {
		return errors.NewWrapErrorWithSuggestions(err, apiKeyUseFailedErrorMsg, fmt.Sprintf(apiKeyUseFailedSuggestions, args[0]))
	}

	output.Printf(c.Config.EnableColor, useAPIKeyMsg, args[0])
	return nil
}

func (c *command) useAPIKey(apiKey, clusterId string) error {
	kcc, err := kafka.FindCluster(c.V2Client, c.Context, clusterId)
	if err != nil {
		return err
	}

	if _, ok := kcc.APIKeys[apiKey]; !ok {
		// check if this is API key exists server-side
		key, httpResp, err := c.V2Client.GetApiKey(apiKey)
		if err != nil {
			return errors.CatchCCloudV2Error(err, httpResp)
		}
		// check if the key is for the right cluster
		if key.Spec.Resource.Id != clusterId {
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(errors.InvalidApiKeyErrorMsg, apiKey, clusterId),
				fmt.Sprintf(errors.InvalidApiKeySuggestions, clusterId),
			)
		}
		// the requested api-key exists, but the secret is not saved locally
		return &errors.UnconfiguredAPISecretError{APIKey: apiKey, ClusterID: clusterId}
	}
	kcc.APIKey = apiKey
	return c.Config.Save()
}
