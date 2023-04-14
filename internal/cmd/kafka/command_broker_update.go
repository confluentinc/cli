package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
)

func (c *brokerCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [broker-id]",
		Short: "Update Kafka broker configurations.",
		Long:  "Update per-broker or cluster-wide Kafka broker configurations.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for broker 1.",
				Code: "confluent kafka broker update 1 --config min.insync.replicas=2,num.partitions=2",
			},
			examples.Example{
				Text: "Update configuration values for all brokers in the cluster.",
				Code: "confluent kafka broker update --all --config min.insync.replicas=2,num.partitions=2",
			},
		),
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the broker being updated.`)
	cmd.Flags().Bool("all", false, "Apply config update to all brokers in the cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *brokerCommand) update(cmd *cobra.Command, args []string) error {
	brokerId, all, err := checkAllOrBrokerIdSpecified(cmd, args)
	if err != nil {
		return err
	}

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.ConfigFlagToMap(configs)
	if err != nil {
		return err
	}
	data := toAlterConfigBatchRequestDataOnPrem(configMap)

	if all {
		resp, err := restClient.ConfigsV3Api.UpdateKafkaClusterConfigs(restContext, clusterId,
			&kafkarestv3.UpdateKafkaClusterConfigsOpts{
				AlterConfigBatchRequestData: optional.NewInterface(data),
			})
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	} else {
		resp, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsalterPost(restContext, clusterId, brokerId,
			&kafkarestv3.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts{
				AlterConfigBatchRequestData: optional.NewInterface(data),
			})
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	}

	if output.GetFormat(cmd) == output.Human {
		if all {
			output.Printf("Updated the following broker configurations for cluster \"%s\":\n", clusterId)
		} else {
			output.Printf("Updated the following configurations for broker \"%d\":\n", brokerId)
		}
	}

	list := output.NewList(cmd)
	for _, config := range data.Data {
		list.Add(&configOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
