package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *brokerCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Short: "List Kafka brokers.",
		Long:  "List Kafka brokers using Confluent Kafka REST.",
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
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "BrokerId", "Host", "Port"}, []string{"Cluster ID", "Broker ID", "Host", "Port"}, []string{"cluster_id", "broker_id", "host", "port"})
	if err != nil {
		return err
	}

	for _, data := range brokersGetResp.Data {
		s := &struct {
			ClusterId string
			BrokerId  int32
			Host      string
			Port      int32
		}{
			ClusterId: data.ClusterId,
			BrokerId:  data.BrokerId,
		}
		if data.Host != nil {
			s.Host = *(data.Host)
		}
		if data.Port != nil {
			s.Port = *(data.Port)
		}
		outputWriter.AddElement(s)
	}

	return outputWriter.Out()
}
