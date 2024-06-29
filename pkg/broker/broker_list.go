package broker

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func List(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	brokers, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersGet(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, broker := range brokers.Data {
		out := &out{
			ClusterId: broker.ClusterId,
			BrokerId:  broker.BrokerId,
		}
		if broker.Host != nil {
			out.Host = *broker.Host
		}
		if broker.Port != nil {
			out.Port = *broker.Port
		}
		list.Add(out)
	}
	return list.Print()
}
