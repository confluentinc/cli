package kafka

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *clusterCommand) newEndpointUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Kafka cluster endpoint.",
		Long:  "Use a Kafka cluster endpoint as active endpoint for all subsequent Kafka cluster commands in current environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.endpointUse,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443" for subsequent Kafka cluster commands.`,
				Code: `confluent kafka cluster endpoint use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443"`,
			},
		),
	}

	return cmd
}

func (c *clusterCommand) endpointUse(cmd *cobra.Command, args []string) error {
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
			"Current active Kafka cluster is empty",
			"Please run `confluent kafka cluster use` to set the active cluster first.",
		)
	}

	endpoint := args[0]

	if valid := validateUserProvidedKafkaClusterEndpoint(endpoint, activeCluster, c); !valid {
		suggestion := `Please run "confluent kafka cluster endpoint list" to see all available Kafka cluster endpoints, or "confluent kafka cluster use" to switch to a different cluster.`
		return errors.NewErrorWithSuggestions(fmt.Sprintf("Kafka cluster endpoint %q is invalid for active cluster %q", endpoint, activeCluster), suggestion)
	}

	c.Context.KafkaClusterContext.SetActiveKafkaClusterEndpoint(endpoint)

	// reset the last update time if we change the active endpoint
	activeClusterConfig := c.Context.KafkaClusterContext.GetKafkaClusterConfig(activeCluster)
	if activeClusterConfig != nil {
		activeClusterConfig.LastUpdate = time.Time{}
	}

	if err := c.Config.Save(); err != nil {
		return err
	}

	return nil
}

func validateUserProvidedKafkaClusterEndpoint(endpoint, activeCluster string, c *clusterCommand) bool {
	// check if the specified endpoint exists in the current active cluster's endpoint list, and error out if not
	activeClusterConfigs, _, err := c.V2Client.DescribeKafkaCluster(activeCluster, c.Context.GetCurrentEnvironment())
	if err != nil {
		log.CliLogger.Debugf("Error describing Kafka Cluster: %v", err)
		return false
	}

	activeClusterEndpoints := activeClusterConfigs.Spec.GetEndpoints()

	for _, v := range activeClusterEndpoints {
		if v.GetHttpEndpoint() == endpoint {
			log.CliLogger.Debugf("The specified endpoint %q is a valid %q endpoint", endpoint, v.GetConnectionType())
			output.ErrPrintf(c.Config.EnableColor, "Set Kafka endpoint \"%s\" as the active endpoint for Kafka cluster \"%s\".\n", endpoint, activeCluster)

			return true
		}
	}

	return false
}
