package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func (c *command) newConnectDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <usm-cluster-id>",
		Short:             "Describe a Connect cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectArgs),
		RunE:              c.describeConnect,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a Confluent Platform Connect cluster with the USM ID usmcc-abc123.",
				Code: "confluent unified-stream-manager connect describe usmcc-abc123",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeConnect(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := c.V2Client.GetUsmConnectCluster(args[0], environmentId)
	if err != nil {
		return err
	}

	onPremToCloudKafkaIdMap, err := c.getOnPremToCloudKafkaIdMap(environmentId)
	if err != nil {
		return err
	}

	usmKafkaClusterId, ok := onPremToCloudKafkaIdMap[cluster.GetKafkaClusterId()]
	if !ok {
		log.CliLogger.Errorf(kafkaClusterNotFoundErrorMsg, cluster.GetKafkaClusterId())
	}

	return printConnectTable(cmd, cluster, usmKafkaClusterId)
}
