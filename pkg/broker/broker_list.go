package broker

import (
	"context"

	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func List(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	brokersGetResp, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersGet(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, data := range brokersGetResp.Data {
		broker := &BrokerOut{
			ClusterId: data.ClusterId,
			BrokerId:  data.BrokerId,
		}
		if data.Host != nil {
			broker.Host = *(data.Host)
		}
		if data.Port != nil {
			broker.Port = *(data.Port)
		}
		list.Add(broker)
	}
	return list.Print()
}
