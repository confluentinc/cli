package broker

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func ConfigurationList(cmd *cobra.Command, args []string, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	brokerId, err := GetId(cmd, args)
	if err != nil {
		return err
	}

	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	var data []*ConfigOut
	brokerConfig, err := getIndividualBrokerConfigs(restClient, restContext, clusterId, brokerId, config)
	if err != nil {
		return err
	}
	data = parseBrokerConfigData(brokerConfig)

	list := output.NewList(cmd)
	for _, entry := range data {
		if output.GetFormat(cmd) == output.Human {
			entry.Name = utils.Abbreviate(entry.Name, AbbreviationLength)
			entry.Value = utils.Abbreviate(entry.Value, AbbreviationLength)
		}
		list.Add(entry)
	}
	return list.Print()
}
