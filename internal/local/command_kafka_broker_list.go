package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *Command) newKafkaBrokerListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local Kafka brokers.",
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
		broker := &broker.BrokerOut{
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
