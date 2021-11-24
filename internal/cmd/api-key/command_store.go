package apikey

import (
	"context"
	"fmt"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store [api-key] [secret]",
		Short: "Store an API key/secret locally to use in the CLI.",
		Long:  longDescription,
		Args:  cobra.MaximumNArgs(2),
		RunE:  pcmd.NewCLIRunE(c.store),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Pass the API key and secret as arguments",
				Code: "confluent api-key store my-key my-secret",
			},
			examples.Example{
				Text: "Get prompted for both the API key and secret",
				Code: "confluent api-key store",
			},
			examples.Example{
				Text: "Get prompted for only the API secret",
				Code: "confluent api-key store my-key",
			},
			examples.Example{
				Text: "Pipe the API secret",
				Code: "confluent api-key store my-key -",
			},
			examples.Example{
				Text: "Provide the API secret in a file",
				Code: "confluent api-key store my-key @my-secret.txt",
			},
		),
	}

	cmd.Flags().String(resourceFlagName, "", "The resource ID of the resource the API key is for.")
	cmd.Flags().BoolP("force", "f", false, "Force overwrite existing secret for this key.")

	return cmd
}

func (c *command) store(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	var cluster *v1.KafkaClusterConfig

	// Attempt to get cluster from --resource flag if set; if that doesn't work,
	// attempt to fall back to the currently active Kafka cluster
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err == nil && clusterId != "" {
		if resourceType != pcmd.KafkaResourceType {
			return errors.Errorf(errors.NonKafkaNotImplementedErrorMsg)
		}
		cluster, err = c.Context.FindKafkaCluster(cmd, clusterId)
		if err != nil {
			return err
		}
	} else {
		cluster, err = c.Context.GetKafkaClusterForCommand(cmd)
		if err != nil {
			return err
		}
	}

	var key string
	if len(args) == 0 {
		key, err = c.parseFlagResolverPromptValue("", "Key: ", false)
		if err != nil {
			return err
		}
	} else {
		key = args[0]
	}

	var secret string
	if len(args) < 2 {
		secret, err = c.parseFlagResolverPromptValue("", "Secret: ", true)
		if err != nil {
			return err
		}
	} else if len(args) == 2 {
		secret, err = c.parseFlagResolverPromptValue(args[1], "", true)
		if err != nil {
			return err
		}
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Check if API key exists server-side
	apiKey, err := c.Client.APIKey.Get(context.Background(), &schedv1.ApiKey{Key: key, AccountId: c.EnvironmentId()})
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.APIKeyNotFoundSuggestions)
	}

	apiKeyIsValidForTargetCluster := false
	for _, lkc := range apiKey.LogicalClusters {
		if lkc.Id == cluster.ID {
			apiKeyIsValidForTargetCluster = true
			break
		}
	}
	if !apiKeyIsValidForTargetCluster {
		return errors.NewErrorWithSuggestions(errors.APIKeyNotValidForClusterErrorMsg, errors.APIKeyNotValidForClusterSuggestions)
	}

	// API key exists server-side... now check if API key exists locally already
	if found, err := c.keystore.HasAPIKey(key, cluster.ID, cmd); err != nil {
		return err
	} else if found && !force {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.RefuseToOverrideSecretErrorMsg, key),
			fmt.Sprintf(errors.RefuseToOverrideSecretSuggestions, key))
	}

	if err := c.keystore.StoreAPIKey(&schedv1.ApiKey{Key: key, Secret: secret}, cluster.ID, cmd); err != nil {
		return errors.Wrap(err, errors.UnableToStoreAPIKeyErrorMsg)
	}
	utils.ErrPrintf(cmd, errors.StoredAPIKeyMsg, key)
	return nil
}

func (c *command) parseFlagResolverPromptValue(source, prompt string, secure bool) (string, error) {
	val, err := c.flagResolver.ValueFrom(source, prompt, secure)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}
