package local

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

func (c *Command) newBrokerUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a Confluent local broker configurations.",
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

	pcmd.AddConfigFlag(cmd)
	cmd.Flags().Bool("all", false, "Apply configuration update to all brokers in the cluster.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *Command) update(cmd *cobra.Command, args []string) error {
	brokerId, all, err := broker.CheckAllOrBrokerIdSpecified(cmd, args)
	if err != nil {
		return err
	}

	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.GetMap(config)
	if err != nil {
		return err
	}
	data := broker.ToAlterConfigBatchRequestDataOnPrem(configMap)

	if all {
		resp, err := restClient.ConfigsV3Api.UpdateKafkaClusterConfigs(context.Background(), clusterId,
			&kafkarestv3.UpdateKafkaClusterConfigsOpts{
				AlterConfigBatchRequestData: optional.NewInterface(data),
			})
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	} else {
		resp, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsalterPost(context.Background(), clusterId, brokerId,
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
		list.Add(&broker.ConfigOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
