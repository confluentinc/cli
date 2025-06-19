package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func (c *clusterCommand) newEndpointUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Kafka Cluster endpoint.",
		Long:  "Use a Kafka Cluster endpoint as active endpoint for all subsequent Kafka Cluster commands in current environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.endpointUse,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443" for subsequent Kafka Cluster commands.`,
				Code: `confluent kafka cluster endpoint use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443"`,
			},
		),
	}

	return cmd
}

func (c *clusterCommand) endpointUse(cmd *cobra.Command, args []string) error {
	// TODO: add logic and endpoint validation here! Sanity checks!

	currEnvironment := c.Context.GetCurrentEnvironment()
	if currEnvironment == "" {
		return errors.NewErrorWithSuggestions(
			"Current environment is empty",
			"Please run `confluent environment use` to set the current environment first.",
		)
	}

	// check if active cluster is set, error out if not
	activeCluster := c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
	if activeCluster == "" {
		return errors.NewErrorWithSuggestions(
			"Current active Kafka vluster is empty",
			"Please run `confluent kafka cluster use` to set the active cluster first.",
		)
	}

	endpoint := args[0]

	if valid := validateUserProvidedKafkaClusterEndpoint(endpoint, activeCluster, c); !valid {
		suggestion := `Please run "confluent kafka cluster endpoint list" to see all available Kafka cluster endpoints, or "confluent kafka cluster use" to switch to a different cluster.`
		return errors.NewErrorWithSuggestions(fmt.Sprintf("Kafka cluster endpoint %q is invalid for active cluster %q", endpoint, activeCluster), suggestion)
	}

	c.Context.KafkaClusterContext.SetActiveKafkaClusterEndpoint(endpoint)

	if err := c.Config.Save(); err != nil {
		return err
	}

	return nil
}

func validateUserProvidedKafkaClusterEndpoint(endpoint, activeCluster string, c *clusterCommand) bool {

	// check if the specified endpoint exists in the endpoint list, error out if not
	// check if the specified endpoint corresponds to active cluster (cloud & region), and error out if not

	activeClusterConfigs, _, err := c.V2Client.DescribeKafkaCluster(activeCluster, c.Context.GetCurrentEnvironment())
	if err != nil {
		log.CliLogger.Debugf("Error describing Kafka Cluster: %v", err)
		return false
	}

	activeClusterEndpoints := activeClusterConfigs.Spec.GetEndpoints()

	for _, v := range activeClusterEndpoints {
		// check how complete this GetHttpEndpoint() returns, we're fine if returns everything.
		// Add additional checks if the GetHttpEndpoint() returns partial endpoints
		if v.GetHttpEndpoint() == endpoint {
			log.CliLogger.Debugf("The specified endpoint %q is a valid %q endpoint", endpoint, v.GetConnectionType())
			return true
		}
	}

	return false
}
