package cmd

import (
	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
)

func GetKafkaCluster(cmd *cobra.Command, ch *ConfigHelper) (*kafkav1.KafkaCluster, error) {
	clusterID, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return nil, err
	}
	environment, err := GetEnvironment(cmd, ch.Config)
	if err != nil {
		return nil, err
	}
	return ch.KafkaCluster(clusterID, environment)
}

func GetKafkaClusterConfig(cmd *cobra.Command, ch *ConfigHelper) (*config.KafkaClusterConfig, error) {
	clusterID, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return nil, err
	}
	environment, err := GetEnvironment(cmd, ch.Config)
	if err != nil {
		return nil, err
	}
	return ch.KafkaClusterConfig(clusterID, environment)
}

func GetEnvironment(cmd *cobra.Command, cfg *config.Config) (string, error) {
	var environment string
	if cmd.Flags().Lookup("environment") != nil {
		var err error
		environment, err = cmd.Flags().GetString("environment")
		if err != nil {
			return "", err
		}
	}
	if environment == "" {
		environment = cfg.Auth.Account.Id
	}
	return environment, nil
}
