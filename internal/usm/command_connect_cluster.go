package usm

import (
	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type connectClusterCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type connectClusterOut struct {
	ID                              string `human:"ID" serialized:"id"`
	ConfluentPlatformConnectCluster string `human:"Confluent Platform Connect Cluster" serialized:"confluent_platform_connect_cluster"`
	KafkaClusterId                  string `human:"Kafka Cluster Id" serialized:"kafka_cluster_id"`
	UsmKafkaClusterId               string `human:"USM Kafka Cluster Id" serialized:"usm_kafka_cluster_id"`
	Environment                     string `human:"Environment" serialized:"environment"`
	Cloud                           string `human:"Cloud" serialized:"cloud"`
	Region                          string `human:"Region" serialized:"region"`
}

func newConnectClusterCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command { //nolint:unparam
	cmd := &cobra.Command{
		Use:         "connect-cluster",
		Aliases:     []string{"connect"},
		Short:       "Manage Confluent Cloud USM Connect clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &connectClusterCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}

	cmd.AddCommand(
		c.newCreateCommand(),
		c.newDeleteCommand(),
		c.newDescribeCommand(),
		c.newListCommand(),
	)

	return cmd
}

func printConnectCluster(cmd *cobra.Command, connectCluster usmv1.UsmV1ConnectCluster) error {
	table := output.NewTable(cmd)
	out := &connectClusterOut{
		ID:                              connectCluster.GetId(),
		ConfluentPlatformConnectCluster: connectCluster.GetConfluentPlatformConnectClusterId(),
		KafkaClusterId:                  connectCluster.GetKafkaClusterId(),
		Environment:                     connectCluster.Environment.GetId(),
		UsmKafkaClusterId:               connectCluster.GetUsmKafkaClusterId(),
		Cloud:                           connectCluster.GetCloud(),
		Region:                          connectCluster.GetRegion(),
	}
	table.Add(out)
	return table.Print()
}

func (c *connectClusterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *connectClusterCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectClusters()
}

func (c *connectClusterCommand) autocompleteConnectClusters() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}
	connectClusters, err := c.V2Client.ListUsmConnectClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(connectClusters))
	for i, connectCluster := range connectClusters {
		suggestions[i] = connectCluster.GetId()
	}
	return suggestions
}
