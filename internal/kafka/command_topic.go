package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

const numPartitionsKey = "num.partitions"

type command struct {
	*pcmd.AuthenticatedCLICommand
	clientID string
}

func newTopicCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage Kafka topics.",
	}

	c := &command{clientID: cfg.Version.ClientID}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newCreateCommandOnPrem())
		cmd.AddCommand(c.newDeleteCommandOnPrem())
		cmd.AddCommand(c.newDescribeCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
		cmd.AddCommand(c.newUpdateCommandOnPrem())
	}

	cmd.AddCommand(c.newConfigurationCommand(cfg))

	return cmd
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *command) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTopics(cmd)
}

func (c *command) autocompleteTopics(cmd *cobra.Command) []string {
	topics, err := c.getTopics(cmd)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(topics))
	for i, topic := range topics {
		var description string
		if topic.GetIsInternal() {
			description = "Internal"
		}
		suggestions[i] = fmt.Sprintf("%s\t%s", topic.GetTopicName(), description)
	}
	return suggestions
}

func (c *command) provisioningClusterCheck(lkc string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, environmentId)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc, httpResp)
	}
	if cluster.Status.Phase == ccloudv2.StatusProvisioning {
		return fmt.Errorf(errors.KafkaRestProvisioningErrorMsg, lkc)
	}
	return nil
}

func addApiKeyToCluster(cmd *cobra.Command, cluster *config.KafkaClusterConfig) error {
	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}

	if apiKey != "" {
		apiSecret, err := cmd.Flags().GetString("api-secret")
		if err != nil {
			return err
		}

		cluster.APIKey = apiKey
		cluster.APIKeys[cluster.APIKey] = &config.APIKeyPair{
			Key:    apiKey,
			Secret: apiSecret,
		}
	}

	if cluster.APIKey == "" {
		return &errors.UnspecifiedAPIKeyError{ClusterID: cluster.ID}
	}

	if pair, ok := cluster.APIKeys[cluster.APIKey]; !ok || pair.Secret == "" {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`no secret for API key "%s" of resource "%s" passed via flag or stored in local CLI state`, apiKey, cluster.ID),
			fmt.Sprintf("Pass the API secret with flag `--api-secret` or store with `confluent api-key store %s --resource %s`.", apiKey, cluster.ID),
		)
	}

	return nil
}
