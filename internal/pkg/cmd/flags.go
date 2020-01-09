package cmd

import (
	"context"
	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
)

func GetKafkaCluster(cmd *cobra.Command, ch *ConfigHelper, flag ...string) (*kafkav1.KafkaCluster, error) {
	if len(flag) == 0 {
		flag = []string{"cluster"}
	}
	clusterID := ""
	if cmd.Flags().Lookup(flag[0]) != nil {
		var err error
		clusterID, err = cmd.Flags().GetString(flag[0])
		if err != nil {
			return nil, err
		}
	}
	environment, err := GetEnvironment(cmd, ch.Config)
	if err != nil {
		return nil, err
	}
	return ch.KafkaCluster(clusterID, environment)
}

func GetKafkaClusterConfig(cmd *cobra.Command, ch *ConfigHelper, flag ...string) (*config.KafkaClusterConfig, error) {
	if len(flag) == 0 {
		flag = []string{"cluster"}
	}
	clusterID, err := cmd.Flags().GetString(flag[0])
	if err != nil {
		return nil, err
	}
	credType, err := ch.Config.CredentialType()
	if err != nil {
		return nil, err
	}
	switch credType {
	case config.APIKey:
		return ch.Config.KafkaClusterConfig()
	case config.Username:
		fallthrough
	default:
		environment, err := GetEnvironment(cmd, ch.Config)
		if err != nil {
			return nil, err
		}
		return ch.KafkaClusterConfig(clusterID, environment)
	}
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

func GetSchemaRegistry(cmd *cobra.Command, ch *ConfigHelper) (*srv1.SchemaRegistryCluster, error) {
	ctx := context.Background()
	resourceID, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, err
	}
	environment, err := GetEnvironment(cmd, ch.Config)
	if err != nil {
		return nil, err
	}
	cluster, err := ch.Client.SchemaRegistry.GetSchemaRegistryCluster(
		ctx, &srv1.SchemaRegistryCluster{
			Id:        resourceID,
			AccountId: environment,
		})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func GetKSQL(cmd *cobra.Command, ch *ConfigHelper) (*ksqlv1.KSQLCluster, error) {
	ctx := context.Background()
	resourceID, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, err
	}
	environment, err := GetEnvironment(cmd, ch.Config)
	if err != nil {
		return nil, err
	}
	cluster, err := ch.Client.KSQL.Describe(
		ctx, &ksqlv1.KSQLCluster{
			Id:        resourceID,
			AccountId: environment,
		})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
