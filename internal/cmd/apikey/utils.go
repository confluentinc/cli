package apikey

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *command) srClusterInfo(cmd *cobra.Command, args []string) (accId string, clusterId string, currentKey string, err error) {
	src, err := pcmd.GetSchemaRegistry(cmd, c.ch)
	if err != nil {
		pcmd.Println(cmd, "Schema Registry not set up")
		return "", "", "", errors.HandleCommon(err, cmd)
	}
	clusterInContext, err := c.config.SchemaRegistryCluster()

	if err != nil {
		currentKey = ""
	} else {
		currentKey = clusterInContext.SrCredentials.Key
	}

	return src.AccountId, src.Id, currentKey, nil
}

func (c *command) kafkaClusterInfo(cmd *cobra.Command, args []string) (accId string, clusterId string, currentKey string, err error) {
	kcc, err := pcmd.GetKafkaClusterConfig(cmd, c.ch, "resource")

	if err != nil {
		return "", "", "", errors.HandleCommon(err, cmd)
	}
	return c.config.Auth.Account.Id, kcc.ID, kcc.APIKey, nil
}
//
//func GetKafkaCluster(cmd *cobra.Command, ch *ConfigHelper) (*kafkav1.KafkaCluster, error) {
//	clusterID, err := cmd.Flags().GetString("cluster")
//	if err != nil {
//		return nil, err
//	}
//	environment, err := GetEnvironment(cmd, ch.Config)
//	if err != nil {
//		return nil, err
//	}
//	return ch.KafkaCluster(clusterID, environment)
//}
//
//func GetKafkaClusterConfig(cmd *cobra.Command, ch *ConfigHelper) (*config.KafkaClusterConfig, error) {
//	clusterID, err := cmd.Flags().GetString("cluster")
//	if err != nil {
//		return nil, err
//	}
//	environment, err := GetEnvironment(cmd, ch.Config)
//	if err != nil {
//		return nil, err
//	}
//	return ch.KafkaClusterConfig(clusterID, environment)
//}
