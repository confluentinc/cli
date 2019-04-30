package cmd

import (
	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
)

func GetKafkaCluster(cmd *cobra.Command, cfg *config.Config) (*kafkav1.KafkaCluster, error) {
	clusterID, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return nil, err
	}
	return cfg.KafkaCluster(clusterID)
}

func GetKafkaClusterConfig(cmd *cobra.Command, cfg *config.Config) (config.KafkaClusterConfig, error) {
	clusterID, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return config.KafkaClusterConfig{}, err
	}
	return cfg.KafkaClusterConfig(clusterID)
}
