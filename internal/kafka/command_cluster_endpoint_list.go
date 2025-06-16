package kafka

import (
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		RunE:  c.endpointList,
		Short: "List Kafka Cluster endpoint.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available Kafka Cluster endpoints with current cloud provider and region.",
				Code: "confluent kafka cluster endpoint list --cluster lkc-123456",
			},
		),
	}

	// have a --cluster flag to display endpoint of one specific cluster (use default active cluster if not specified)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

/*
cmk SDK
- name: endpoints
        $ref: '#/components/schemas/cmk.v2.EndpointsMap'
        description: |
          A map of endpoints for connecting to the Kafka cluster,
          keyed by access_point_id. Access Point ID 'public' and 'privatelink' are reserved.
          These can be used for different network access methods or regions.
        read_only: true
        example:
          "ap1pni123":
            kafka_bootstrap_endpoint: "lkc-s1232-00000.us-central1.gcp.private.confluent.cloud:9092"
            http_endpoint: "https://lkc-s1232.us-central1.gcp.private.confluent.cloud:443"
            connection_type: "PRIVATENETWORKINTERFACE"
          "ap2platt67890":
            kafka_bootstrap_endpoint: "lkc-00000-00000.us-central1.gcp.glb.confluent.cloud:9092"
            http_endpoint: "https://lkc-00000-00000.us-central1.gcp.glb.confluent.cloud"
            connection_type: "PRIVATELINK"
*/

func (c *command) endpointList(cmd *cobra.Command, args []string) error {
	// Add logic for displaying endpoint, layer by layer
	// check current environment... cloud... region...
	// PNI -> privatelink -> public

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	if cluster == "" {
		cluster = c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
		if cluster == "" {
			return fmt.Errorf("No cluster specified and no active cluster found, please specify a Kafka cluster Id.")
		}
	}

	// display the endpoints corresponding to the specified cluster

	return nil
}
