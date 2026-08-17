package usm

import (
	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

// connectClusterCloudOut is the single-object output shape for a Connect cluster
// backed by a Confluent Cloud (lkc-) Kafka cluster. It relabels the generic
// kafka_cluster_id as "Kafka Cluster ID" instead of "Confluent Platform Kafka Cluster Id".
type connectClusterCloudOut struct {
	ID                              string `human:"ID" serialized:"id"`
	ConfluentPlatformConnectCluster string `human:"Confluent Platform Connect Cluster" serialized:"confluent_platform_connect_cluster"`
	UsmKafkaClusterId               string `human:"USM Kafka Cluster Id" serialized:"usm_kafka_cluster_id"`
	KafkaClusterId                  string `human:"Kafka Cluster Id" serialized:"kafka_cluster_id"`
	Environment                     string `human:"Environment" serialized:"environment"`
	Cloud                           string `human:"Cloud" serialized:"cloud"`
	Region                          string `human:"Region" serialized:"region"`
}

// isCloudKafkaCluster reports whether the metadata Kafka cluster id refers to a
// Confluent Cloud Kafka cluster (lkc-...), as opposed to a Confluent Platform one.
func isCloudKafkaCluster(kafkaClusterId string) bool {
	return resource.LookupType(kafkaClusterId) == resource.KafkaCluster
}

// printConnectClusterByType prints a single Connect cluster, choosing the output
// shape based on whether its metadata Kafka cluster is Confluent Cloud or Confluent Platform.
func printConnectClusterByType(cmd *cobra.Command, connectCluster usmv1.UsmV1ConnectCluster) error {
	if !isCloudKafkaCluster(connectCluster.GetKafkaClusterId()) {
		return printConnectCluster(cmd, connectCluster)
	}

	table := output.NewTable(cmd)
	table.Add(&connectClusterCloudOut{
		ID:                              connectCluster.GetId(),
		ConfluentPlatformConnectCluster: connectCluster.GetConfluentPlatformConnectClusterId(),
		UsmKafkaClusterId:               connectCluster.GetUsmKafkaClusterId(),
		KafkaClusterId:                  connectCluster.GetKafkaClusterId(),
		Environment:                     connectCluster.Environment.GetId(),
		Cloud:                           connectCluster.GetCloud(),
		Region:                          connectCluster.GetRegion(),
	})
	return table.Print()
}
