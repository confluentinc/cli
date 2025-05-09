package apikey

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

const longDescription = `Use this command to register an API secret created by another process and store it locally.

When you create an API key with the CLI, it is automatically stored locally.
However, when you create an API key using the UI, API, or with the CLI on another
machine, the secret is not available for CLI use until you "store" it. This is because
secrets are irretrievable after creation.

You must have an API secret stored locally for certain CLI commands to
work. For example, the Kafka topic consume and produce commands require an API secret.`

func (c *command) newStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store <api-key> <secret>",
		Short: "Store an API key/secret locally to use in the CLI.",
		Long:  longDescription,
		Args:  cobra.ExactArgs(2),
		RunE:  c.store,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Pass the API key and secret as arguments",
				Code: "confluent api-key store my-key my-secret",
			},
		),
	}

	c.addResourceFlag(cmd, true)
	cmd.Flags().BoolP("force", "f", false, "Force overwrite existing secret for this key.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) store(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	var cluster *config.KafkaClusterConfig

	// Attempt to get cluster from --resource flag if set; if that doesn't work,
	// attempt to fall back to the currently active Kafka cluster
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.V2Client)
	if err == nil && clusterId != "" {
		if resourceType != resource.KafkaCluster {
			return errors.New(nonKafkaNotImplementedErrorMsg)
		}
		cluster, err = kafka.FindCluster(c.V2Client, c.Context, clusterId)
		if err != nil {
			return err
		}
	} else {
		cluster, err = kafka.GetClusterForCommand(c.V2Client, c.Context)
		if err != nil {
			// Replace the error msg since it suggests flags which are unavailable with this command
			return errors.NewErrorWithSuggestions(
				errors.NoKafkaSelectedErrorMsg,
				apiKeyNotValidForClusterSuggestions,
			)
		}
	}

	key := args[0]
	secret := args[1]

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Check if API key exists server-side
	apiKey, httpResp, err := c.V2Client.GetApiKey(key)
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, getOperation, httpResp)
	}

	apiKeyIsValidForTargetCluster := cluster.GetId() != "" && cluster.GetId() == apiKey.GetSpec().Resource.GetId()

	if !apiKeyIsValidForTargetCluster {
		return errors.NewErrorWithSuggestions(
			"the provided API key does not belong to the target cluster",
			apiKeyNotValidForClusterSuggestions,
		)
	}

	// API key exists server-side... now check if API key exists locally already
	if found, err := c.keystore.HasAPIKey(c.V2Client, key, cluster.ID); err != nil {
		return err
	} else if found && !force {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`refusing to overwrite existing secret for API Key "%s"`, key),
			fmt.Sprintf(refuseToOverrideSecretSuggestions, key),
		)
	}

	if err := c.keystore.StoreAPIKey(c.V2Client, &config.APIKeyPair{Key: key, Secret: secret}, cluster.ID); err != nil {
		return fmt.Errorf(unableToStoreApiKeyErrorMsg, err)
	}

	output.ErrPrintf(c.Config.EnableColor, "Stored secret for API key \"%s\".\n", key)
	return nil
}
