package kafka

import (
	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"sort"
	"strconv"
)

type brokerCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewBrokerCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	brokerCmd := &brokerCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:   "broker",
				Short: "Manage Kafka brokers.",
			}, prerunner, OnPremTopicSubcommandFlags),
	}
	brokerCmd.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(brokerCmd.AuthenticatedCLICommand))
	brokerCmd.init()
	return brokerCmd.Command
}

func (brokerCmd *brokerCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(brokerCmd.list),
		Short: "List Kafka brokers.",
		// TODO example
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <broker>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.describe),
		Short: "Describe a Kafka broker.",
		Long:  "See broker configurations and partition-replica information for the spcified broker.",
		// TODO example
	}
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.update),
		Short: "Update Kafka broker.",
		// TODO example
	}
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of configuration overrides ('key=value') for the broker being updated.")
	updateCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	updateCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	updateCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(updateCmd)
}

func (brokerCmd *brokerCommand) list(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(brokerCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Get Brokers
	brokersGetResp, resp, err := restClient.BrokerApi.ClustersClusterIdBrokersGet(restContext, clusterId)
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
			Host:      *(data.Host),
			Port:      *(data.Port),
		}
		outputWriter.AddElement(s)
	}
	return outputWriter.Out()
}

type BrokerData struct {
	BrokerConfigData []kafkarestv3.BrokerConfigData `json:"config_data" yaml:"config_data"`
	ReplicaData      []kafkarestv3.ReplicaData      `json:"replica_data" yaml:"replica_data"`
}

func (brokerCmd *brokerCommand) describe(cmd *cobra.Command, args []string) error {
	brokerIdStr := args[0]
	i, err := strconv.ParseInt(brokerIdStr, 10, 32)
	if err != nil {
		return err
	}
	brokerId := int32(i)
	format, err := cmd.Flags().GetString(output.FlagName)
	restClient, restContext, err := initKafkaRest(brokerCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Get Broker Configs
	brokerConfigsResp, resp, err := restClient.ConfigsApi.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, brokerId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	// Get partition-replicas
	partitionReplicaResp, resp, err := restClient.BrokerApi.ClustersClusterIdBrokersBrokerIdPartitionReplicasGet(restContext, clusterId, brokerId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	brokerInfo := &BrokerData{
		BrokerConfigData: brokerConfigsResp.Data,
		ReplicaData: 	  partitionReplicaResp.Data,
	}
	return output.StructuredOutputForCommand(cmd, format, brokerInfo)
}

func (brokerCmd *brokerCommand) update(cmd *cobra.Command, args []string) error {
	brokerIdStr := args[0]
	i, err := strconv.ParseInt(brokerIdStr, 10, 32)
	if err != nil {
		return err
	}
	brokerId := int32(i)
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if !output.IsValidFormatString(format) { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}
	restClient, restContext, err := initKafkaRest(brokerCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// TODO factor config parsing code out -- shared in updateTopic func
	configStrings, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configsMap, err := kafka.ToMap(configStrings)
	if err != nil {
		return err
	}
	configs := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	j := 0
	for k, v := range configsMap {
		v2 := v
		configs[j] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:      k,
			Value:     &v2,
			Operation: nil,
		}
		j++
	}
	resp, err := restClient.ConfigsApi.ClustersClusterIdBrokersBrokerIdConfigsalterPost(restContext, clusterId, brokerId,
		&kafkarestv3.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts{
			AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs}),
		})
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	if format == output.Human.String() {
		// no errors (config update successful)
		utils.Printf(cmd, "Updated the following configs for broker \"%d\":\n", brokerId)
		// Print Updated Configs
		tableLabels := []string{"Name", "Value"}
		tableEntries := make([][]string, len(configs))
		for i, config := range configs {
			tableEntries[i] = printer.ToRow(
				&struct {
					Name  string
					Value string
				}{Name: config.Name, Value: *config.Value}, []string{"Name", "Value"})
		}
		sort.Slice(tableEntries, func(i int, j int) bool {
			return tableEntries[i][0] < tableEntries[j][0]
		})
		printer.RenderCollectionTable(tableEntries, tableLabels)
	} else { //json or yaml
		sort.Slice(configs, func(i int, j int) bool {
			return configs[i].Name < configs[j].Name
		})
		err = output.StructuredOutput(format, configs)
		if err != nil {
			return err
		}
	}
	return nil
}
