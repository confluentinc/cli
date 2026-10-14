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

	resourceType, clusterId, _, resolveErr := c.resolveResourceId(cmd, c.V2Client)
	isGlobalKey := apiKey.GetSpec().Resource.GetKind() == "Global"

	// Detect Global API keys by either the explicit --resource global flag or the server-side resource Kind.
	// Global keys are org-scoped and stored separately from cluster-scoped keys.
	if resourceType == resource.Global || isGlobalKey {
		if resourceType == resource.Global && !isGlobalKey {
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf("API key %q is not a Global API key", key),
				"Omit `--resource global`, or pass the correct resource ID.",
			)
		}
		if isGlobalKey && resourceType != "" && resourceType != resource.Global {
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf("API key %q is a Global API key, but --resource was set to %q", key, resourceType),
				"Re-run with `--resource global`, or omit `--resource` to auto-detect.",
			)
		}
		return c.storeGlobal(key, secret, force)
	}

	var cluster *config.KafkaClusterConfig

	// Attempt to get cluster from --resource flag if set; if that doesn't work,
	// attempt to fall back to the currently active Kafka cluster.
	// Preserve historical behavior: if --resource resolution failed, silently fall through to the
	// active cluster (the cluster/key-mismatch check below will surface real problems).
	if resolveErr == nil && clusterId != "" {
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

func (c *command) storeGlobal(key, secret string, force bool) error {
	if c.keystore.HasGlobalAPIKey(key) && !force {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`refusing to overwrite existing secret for API Key "%s"`, key),
			fmt.Sprintf(refuseToOverrideSecretSuggestions, key),
		)
	}
	if err := c.keystore.StoreGlobalAPIKey(&config.APIKeyPair{Key: key, Secret: secret}); err != nil {
		return fmt.Errorf(unableToStoreApiKeyErrorMsg, err)
	}
	output.ErrPrintf(c.Config.EnableColor, "Stored secret for Global API key \"%s\".\n", key)
	return nil
}
