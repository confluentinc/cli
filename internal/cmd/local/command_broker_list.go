package local

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type brokerOut struct {
	ClusterId string `human:"Cluster" serialized:"cluster_id"`
	BrokerId  int32  `human:"Broker ID" serialized:"broker_id"`
	Host      string `human:"Host" serialized:"host"`
	Port      int32  `human:"Port" serialized:"port"`
}

func (c *Command) newBrokerListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local cluster brokers.",
		Args:  cobra.NoArgs,
		RunE:  c.brokerList,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *Command) brokerList(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	// Get Brokers
	brokersGetResp, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersGet(context.Background(), clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, data := range brokersGetResp.Data {
		broker := &brokerOut{
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
