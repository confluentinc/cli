package ccloudv2

import (
	"context"
	kafkarest_cc "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"os"
)

type CloudKafkaRESTProvider func() (*CloudKafkaREST, error)

type CloudKafkaREST struct {
	Client  *kafkarest_cc.APIClient
	Context context.Context
}

func getKafkaRestEndpoint(ctx *pcmd.DynamicContext) (string, string, error) {
	if os.Getenv("XX_CCLOUD_USE_KAFKA_API") != "" {
		return "", "", nil
	}
	clusterConfig, err := ctx.GetKafkaClusterForCommand()
	if err != nil {
		return "", "", err
	}
	if clusterConfig.RestEndpoint != "" {
		return clusterConfig.RestEndpoint, clusterConfig.ID, nil
	}
	// if clusterConfig.RestEndpoint is empty, fetch the cluster to ensure config isn't just out of date
	// potentially remove this once Rest Proxy is enabled across prod
	client := pcmd.NewContextClient(ctx)
	kafkaCluster, err := client.FetchCluster(clusterConfig.ID)
	if err != nil {
		return "", clusterConfig.ID, err
	}
	// no need to update the config if it's still empty
	if kafkaCluster.RestEndpoint == "" {
		return "", clusterConfig.ID, nil
	}
	// update config to have updated cluster if rest endpoint is no longer ""
	refreshedClusterConfig := KafkaClusterToKafkaClusterConfig(kafkaCluster)
	ctx.KafkaClusterContext.AddKafkaClusterConfig(refreshedClusterConfig)
	err = ctx.Save() //should we fail on this error or log and continue?
	if err != nil {
		return "", clusterConfig.ID, err
	}
	return kafkaCluster.RestEndpoint, clusterConfig.ID, nil
}
