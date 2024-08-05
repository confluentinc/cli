package broker

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func Describe(cmd *cobra.Command, args []string, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	brokerId, err := GetId(cmd, args)
	if err != nil {
		return err
	}

	broker, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersBrokerIdGet(restContext, clusterId, brokerId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
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
	table.Add(out)

	return table.Print()
}
