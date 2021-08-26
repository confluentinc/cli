package kafka

import (
	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"net/http"
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
		//Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.describe),
		Short: "Describe a Kafka broker.",
		Long:  "See cluster-wide or per broker configuration values.",
		// TODO example
	}
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().Bool("all", false,"Get cluster-wide broker configurations (non-default values only).")
	describeCmd.Flags().Int32("broker", -1,"Get configuration values for specific broker ID.")
	describeCmd.Flags().String("config-name", "", "Get a specific configuration value (pair with --all to see a a cluster-wide config.")
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

type ConfigData struct {
	Name        string              `json:"name"`
	Value       string             `json:"value,omitempty"`
	IsDefault   bool                `json:"is_default"`
	IsReadOnly  bool                `json:"is_read_only"`
	IsSensitive bool                `json:"is_sensitive"`
}

func (brokerCmd *brokerCommand) describe(cmd *cobra.Command, args []string) error {
	// TODO config name flag
	brokerId, all, err := checkAllOrBrokerIdSpecified(cmd)
	if err != nil {
		return err
	}
	configName, err := cmd.Flags().GetString("config-name")
	if err != nil {
		return err
	}
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
	var data []ConfigData
	if all {
		var clusterConfig kafkarestv3.ClusterConfigDataList
		var resp *http.Response
		var err error
		if configName != "" {
			var configNameData kafkarestv3.ClusterConfigData
			configNameData, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokerConfigsNameGet(restContext, clusterId, configName)
			clusterConfig.Data = []kafkarestv3.ClusterConfigData{configNameData}
		} else {
			clusterConfig, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokerConfigsGet(restContext, clusterId)
		}

		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		}
		data = parseClusterConfigData(clusterConfig)
	} else {
		var brokerConfig kafkarestv3.BrokerConfigDataList
		var resp *http.Response
		var err error
		if configName != "" {
			var brokerNameData kafkarestv3.BrokerConfigData
			brokerNameData, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokersBrokerIdConfigsNameGet(restContext, clusterId, brokerId, configName)
			brokerConfig.Data = []kafkarestv3.BrokerConfigData{brokerNameData}
		} else {
			brokerConfig, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, brokerId)
		}
		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		}
		data = parseBrokerConfigData(brokerConfig)
	}
	if format == output.Human.String() {
		configsTableLabels := []string{"Name", "Value", "Is Default", "Is Read Only", "Is Sensitive"}
		configsTableEntries := make([][]string, len(data))
		for i, entry := range data {
			configsTableEntries[i] = printer.ToRow(&struct {
				name  		string
				value 		string
				isDefault	bool
				isReadOnly  bool
				isSensitive bool
			}{name: utils.Abbreviate(entry.Name, 30), value: utils.Abbreviate(entry.Value, 20), isDefault: entry.IsDefault, isReadOnly: entry.IsReadOnly, isSensitive: entry.IsSensitive},
			[]string{"name", "value", "isDefault", "isReadOnly", "isSensitive"})
		}
		sort.Slice(configsTableEntries, func(i int, j int) bool {
			return configsTableEntries[i][0] < configsTableEntries[j][0]
		})
		printer.RenderCollectionTable(configsTableEntries, configsTableLabels)
	} else {
		return output.StructuredOutputForCommand(cmd, format, data)
	}
	return nil
}

func parseClusterConfigData(clusterConfig kafkarestv3.ClusterConfigDataList) []ConfigData {
	var configs []ConfigData
	for _, data := range clusterConfig.Data {
		config := ConfigData{
			Name:        data.Name,
			IsDefault:   data.IsDefault,
			IsReadOnly:  data.IsReadOnly,
			IsSensitive: data.IsSensitive,
		}
		if data.Value != nil {
			config.Value = *data.Value
		}
		configs = append(configs, config)
	}
	return configs
}

func parseBrokerConfigData(brokerConfig kafkarestv3.BrokerConfigDataList) []ConfigData {
	var configs []ConfigData
	for _, data := range brokerConfig.Data {
		config := ConfigData{
			Name:        data.Name,
			IsDefault:   data.IsDefault,
			IsReadOnly:  data.IsReadOnly,
			IsSensitive: data.IsSensitive,
		}
		if data.Value != nil {
			config.Value = *data.Value
		} else {
			config.Value = ""
		}
		configs = append(configs, config)
	}
	return configs
}

func checkAllOrBrokerIdSpecified(cmd *cobra.Command) (int32, bool, error) {
	if cmd.Flags().Changed("all") && cmd.Flags().Changed("broker") {
		return -1, false, errors.New("only specify one of these")
	}
	if !cmd.Flags().Changed("all") && !cmd.Flags().Changed("broker") {
		return -1, false, errors.New("must specify one of these")
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return -1, false, err
	}
	brokerId, err := cmd.Flags().GetInt32("broker")
	if err != nil {
		return -1, false, err
	}
	return brokerId, all, nil
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
