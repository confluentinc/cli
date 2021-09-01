package kafka

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/antihax/optional"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type brokerCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type configData struct {
	Name        string `json:"name" yaml:"name"`
	Value       string `json:"value,omitempty" yaml:"value,omitempty"`
	IsDefault   bool   `json:"is_default" yaml:"is_default"`
	IsReadOnly  bool   `json:"is_read_only" yaml:"is_read_only"`
	IsSensitive bool   `json:"is_sensitive" yaml:"is_sensitive"`
}

const abbreviationLength = 25

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
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.describe),
		Short: "Describe a Kafka broker.",
		Long:  "See cluster-wide or per broker configuration values.",
		// TODO example
	}
	describeCmd.Flags().Bool("all", false, "Get cluster-wide broker configurations (non-default values only).")
	describeCmd.Flags().String("config-name", "", "Get a specific configuration value (pair with --all to see a a cluster-wide config.")
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update <broker>",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.update),
		Short: "Update Kafka an broker or cluster-wide broker configs.",
		// TODO example
	}
	updateCmd.Flags().Bool("all", false, "Apply config update to all brokers in the cluster.")
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of configuration overrides ('key=value') for the broker being updated.")
	updateCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	updateCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	check(updateCmd.MarkFlagRequired("config"))
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

func (brokerCmd *brokerCommand) describe(cmd *cobra.Command, args []string) error {
	brokerId, all, err := checkAllOrBrokerIdSpecified(cmd, args)
	if err != nil {
		return err
	}
	configName, err := cmd.Flags().GetString("config-name")
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(brokerCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Get Broker Configs
	var data []configData
	var resp *http.Response
	if all {
		var clusterConfig kafkarestv3.ClusterConfigDataList
		if configName != "" { // Get configName config
			var configNameData kafkarestv3.ClusterConfigData
			configNameData, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokerConfigsNameGet(restContext, clusterId, configName)
			clusterConfig.Data = []kafkarestv3.ClusterConfigData{configNameData}
		} else { // Get all configs
			clusterConfig, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokerConfigsGet(restContext, clusterId)
		}
		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		}
		data = parseClusterConfigData(clusterConfig)
	} else {
		var brokerConfig kafkarestv3.BrokerConfigDataList
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
			entry.Name = utils.Abbreviate(entry.Name, abbreviationLength)
			entry.Value = utils.Abbreviate(entry.Value, abbreviationLength)
			configsTableEntries[i] = printer.ToRow(&entry, []string{"Name", "Value", "IsDefault", "IsReadOnly", "IsSensitive"})
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

func (brokerCmd *brokerCommand) update(cmd *cobra.Command, args []string) error {
	brokerId, all, err := checkAllOrBrokerIdSpecified(cmd, args)
	if err != nil {
		return err
	}
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
	configStrings, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configsMap, err := utils.ToMap(configStrings)
	if err != nil {
		return err
	}
	configs := toAlterConfigBatchRequestData(configsMap)
	if err != nil {
		return err
	}
	if all {
		resp, err := restClient.ConfigsApi.ClustersClusterIdBrokerConfigsalterPost(restContext, clusterId,
			&kafkarestv3.ClustersClusterIdBrokerConfigsalterPostOpts{
				AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs}),
			})
		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		}
	} else {
		resp, err := restClient.ConfigsApi.ClustersClusterIdBrokersBrokerIdConfigsalterPost(restContext, clusterId, brokerId,
			&kafkarestv3.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts{
				AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs}),
			})
		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		}
	}
	if format == output.Human.String() {
		// no errors (config update successful)
		if all {
			utils.Printf(cmd, "Updated the following broker configs for cluster \"%s\":\n", clusterId)
		} else {
			utils.Printf(cmd, "Updated the following configs for broker \"%d\":\n", brokerId)
		}
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

func checkAllOrBrokerIdSpecified(cmd *cobra.Command, args []string) (int32, bool, error) {
	if cmd.Flags().Changed("all") && len(args) > 0 {
		return -1, false, errors.New(errors.OnlySpecifyAllOrBrokerIDErrorMsg)
	}
	if !cmd.Flags().Changed("all") && len(args) == 0 {
		return -1, false, errors.New(errors.MustSpecifyAllOrBrokerIDErrorMsg)
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return -1, false, err
	}
	if len(args) > 0 {
		brokerIdStr := args[0]
		i, err := strconv.ParseInt(brokerIdStr, 10, 32)
		if err != nil {
			return -1, false, err
		}
		brokerId := int32(i)
		return brokerId, false, nil
	}
	return -1, all, nil
}

func parseClusterConfigData(clusterConfig kafkarestv3.ClusterConfigDataList) []configData {
	var configs []configData
	for _, data := range clusterConfig.Data {
		config := configData{
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

func parseBrokerConfigData(brokerConfig kafkarestv3.BrokerConfigDataList) []configData {
	var configs []configData
	for _, data := range brokerConfig.Data {
		config := configData{
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
