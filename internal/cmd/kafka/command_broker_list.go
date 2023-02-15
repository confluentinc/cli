package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type brokerOut struct {
	ClusterId string `human:"Cluster" serialized:"cluster_id"`
	BrokerId  int32  `human:"Broker ID" serialized:"broker_id"`
	Host      string `human:"Host" serialized:"host"`
	Port      int32  `human:"Port" serialized:"port"`
}

func (c *brokerCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka brokers.",
		Long:  "List Kafka brokers using Confluent Kafka REST.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *brokerCommand) list(cmd *cobra.Command, _ []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	// Get Brokers
	brokersGetResp, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersGet(restContext, clusterId)
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
