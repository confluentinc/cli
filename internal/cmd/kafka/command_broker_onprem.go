package kafka

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/confluentinc/cli/internal/pkg/examples"

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

type brokerTaskData struct {
	ClusterId string `json:"cluster_id" yaml:"cluster_id"`
	BrokerId  int32 `json:"broker_id" yaml:"broker_id"`
	TaskType  kafkarestv3.BrokerTaskType `json:"task_type" yaml:"task_type"`
	TaskStatus string `json:"task_status" yaml:"task_status"`
	CreatedAt	time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt	time.Time `json:"updated_at" yaml:"updated_at"`
	ShutdownScheduled bool `json:"shutdown_scheduled,omitempty" yaml:"shutdown_scheduled,omitempty"`
	SubTaskStatuses  string `json:"sub_task_statuses" yaml:"sub_task_statuses"`
	ErrorCode	int32 `json:"error_code,omitempty" yaml:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty" yaml:"error_message,omitempty"`
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
		Long:  "List Kafka brokers using Confluent Kafka REST.",
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe [broker-id]",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.describe),
		Short: "Describe a Kafka broker.",
		Long:  "Describe cluster-wide or per-broker configuration values using Confluent Kafka REST.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the `min.insync.replicas` configuration for broker 1.",
				Code: "confluent kafka broker describe 1 --config-name min.insync.replicas",
			},
			examples.Example{
				Text: "Describe the non-default cluster-wide broker configuration values.",
				Code: "confluent kafka broker describe --all",
			},
		),
	}
	describeCmd.Flags().Bool("all", false, "Get cluster-wide broker configurations (non-default values only).")
	describeCmd.Flags().String("config-name", "", "Get a specific configuration value (pair with --all to see a a cluster-wide config.")
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update [broker-id]",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.update),
		Short: "Update per-broker or cluster-wide Kafka broker configs.",
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
	updateCmd.Flags().Bool("all", false, "Apply config update to all brokers in the cluster.")
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of configuration overrides ('key=value') for the broker being updated.")
	updateCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	updateCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	check(updateCmd.MarkFlagRequired("config"))
	updateCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <broker-id>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.delete),
		Short: "Delete a Kafka broker.",
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	deleteCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(deleteCmd)

	tasksCmd := &cobra.Command{
		Use:   "get-tasks [broker-id]",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(brokerCmd.getTasks),
		Short: "List broker tasks.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List remove-broker tasks for broker 1.",
				Code: "confluent kafka broker get-tasks 1 --task-type remove-broker",
			},
			examples.Example{
				Text: "List broker tasks for all brokers in the cluster",
				Code: "confluent kafka broker get-tasks --all",
			},
		),
	}
	tasksCmd.Flags().Bool("all", false, "List broker tasks for the cluster.")
	tasksCmd.Flags().String("task-type", "", "Search by task type (add-broker or remove-broker).")
	tasksCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	tasksCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	tasksCmd.Flags().SortFlags = false
	brokerCmd.AddCommand(tasksCmd)
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
	if all { // fetch cluster-wide configs
		clusterConfig, err := getClusterWideConfigs(restClient, restContext, clusterId, configName)
		if err != nil {
			return err
		}
		data = parseClusterConfigData(clusterConfig)
	} else { // fetch individual broker configs
		brokerConfig, err := getIndividualBrokerConfigs(restClient, restContext, clusterId, brokerId, configName)
		if err != nil {
			return err
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
		sort.Slice(configsTableEntries, func(i, j int) bool {
			return configsTableEntries[i][0] < configsTableEntries[j][0]
		})
		printer.RenderCollectionTable(configsTableEntries, configsTableLabels)
	} else {
		return output.StructuredOutputForCommand(cmd, format, data)
	}
	return nil
}

// fetch per-broker configs or just configName config if specified
func getIndividualBrokerConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, brokerId int32, configName string) (kafkarestv3.BrokerConfigDataList, error) {
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
		return brokerConfig, kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	return brokerConfig, nil
}

// fetch cluster-wide configs or just configName config if specified
func getClusterWideConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, configName string) (kafkarestv3.ClusterConfigDataList, error) {
	var clusterConfig kafkarestv3.ClusterConfigDataList
	var resp *http.Response
	var err error
	if configName != "" { // Get configName config
		var configNameData kafkarestv3.ClusterConfigData
		configNameData, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokerConfigsNameGet(restContext, clusterId, configName)
		clusterConfig.Data = []kafkarestv3.ClusterConfigData{configNameData}
	} else { // Get all configs
		clusterConfig, resp, err = restClient.ConfigsApi.ClustersClusterIdBrokerConfigsGet(restContext, clusterId)
	}
	if err != nil {
		return clusterConfig, kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	return clusterConfig, nil
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

func (brokerCmd *brokerCommand) delete(cmd *cobra.Command, args []string) error {
	brokerIdStr := args[0]
	i, err := strconv.ParseInt(brokerIdStr, 10, 32)
	if err != nil {
		return err
	}
	brokerId := int32(i)
	restClient, restContext, err := initKafkaRest(brokerCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	opts := kafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts{
		ShouldShutdown: optional.NewBool(true),
	}
	_, resp, err := restClient.BrokerApi.ClustersClusterIdBrokersBrokerIdDelete(restContext, clusterId, brokerId, &opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	fmt.Printf("Started deletion of broker %d. To monitor the remove-broker task run `confluent kafka broker get-tasks %d --task-type remove-broker`", brokerId, brokerId)
	return nil
}

func (brokerCmd *brokerCommand) getTasks(cmd *cobra.Command, args []string) error {
	brokerId, all, err := checkAllOrBrokerIdSpecified(cmd, args)
	if err != nil {
		return err
	}
	taskName, err := cmd.Flags().GetString("task-type")
	if err != nil {
		return err
	}
	taskType, err := getBrokerTaskType(taskName)
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
	var taskData kafkarestv3.BrokerTaskDataList
	if all { // get BrokerTasks for the cluster
		taskData, err = getBrokerTasksForCluster(restClient, restContext, clusterId, taskType)
		if err != nil {
			return err
		}
	} else { // fetch individual broker configs
		taskData, err = getBrokerTasksForBroker(restClient, restContext, clusterId, brokerId, taskType)
		if err != nil {
			return err
		}
	}
	if format == output.Human.String() {
		configsTableLabels := []string{"Cluster ID", "Broker ID", "Task Type", "Task Status", "Created At", "Updated At", "Shutdown Scheduled", "Subtask Statuses", "Error Code", "Error Message"}
		configsTableEntries := make([][]string, len(taskData.Data))
		for i, entry := range taskData.Data {
			s := parseBrokerTaskData(entry)
			configsTableEntries[i] = printer.ToRow(&s, []string{"ClusterId", "BrokerId", "TaskType", "TaskStatus", "CreatedAt", "UpdatedAt", "ShutdownScheduled", "SubTaskStatuses", "ErrorCode", "ErrorMessage"})
		}
		sort.Slice(configsTableEntries, func(i, j int) bool {
			return configsTableEntries[i][0] < configsTableEntries[j][0]
		})
		printer.RenderCollectionTable(configsTableEntries, configsTableLabels)
	} else {
		printData := make([]brokerTaskData, len(taskData.Data))
		for i, entry := range taskData.Data {
			printData[i] = parseBrokerTaskData(entry)
		}
		return output.StructuredOutputForCommand(cmd, format, printData)
	}
	return nil
}

func parseBrokerTaskData(entry kafkarestv3.BrokerTaskData) brokerTaskData {
	s := brokerTaskData {
		ClusterId: entry.ClusterId,
		BrokerId:  entry.BrokerId,
		TaskType:  entry.TaskType,
		TaskStatus: entry.TaskStatus,
		CreatedAt:  entry.CreatedAt,
		UpdatedAt:  entry.UpdatedAt,
		SubTaskStatuses: mapToKeyValueString(entry.SubTaskStatuses),
	}
	if entry.ShutdownScheduled != nil {
		s.ShutdownScheduled = *entry.ShutdownScheduled
	}
	if entry.ErrorCode != nil {
		s.ErrorCode = *entry.ErrorCode
	}
	if entry.ErrorMessage != nil {
		s.ErrorMessage = *entry.ErrorMessage
	}
	return s
}

func mapToKeyValueString(values map[string]string) string {
	kvString := ""
	for k, v := range values {
		if len(kvString) == 0 {
			kvString = k + "=" + v
		} else {
			kvString = kvString + "\n" + k + "=" + v
		}
	}
	return kvString
}

func getBrokerTasksForCluster(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, taskType kafkarestv3.BrokerTaskType) (kafkarestv3.BrokerTaskDataList, error){
	var taskData kafkarestv3.BrokerTaskDataList
	var resp *http.Response
	var err error
	if taskType != "" {
		taskData, resp, err = restClient.BrokerTaskApi.ClustersClusterIdBrokersTasksTaskTypeGet(restContext, clusterId, taskType)
	} else {
		taskData, resp, err = restClient.BrokerTaskApi.ClustersClusterIdBrokersTasksGet(restContext, clusterId)
	}
	if err != nil {
		return taskData, kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	return taskData, nil
}

func getBrokerTasksForBroker(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, brokerId int32, taskType kafkarestv3.BrokerTaskType) (kafkarestv3.BrokerTaskDataList, error){
	var taskData kafkarestv3.BrokerTaskDataList
	var resp *http.Response
	var err error
	if taskType != "" {
		var brokerTaskData kafkarestv3.BrokerTaskData
		brokerTaskData, resp, err = restClient.BrokerTaskApi.ClustersClusterIdBrokersBrokerIdTasksTaskTypeGet(restContext, clusterId, brokerId, taskType)
		taskData.Data = []kafkarestv3.BrokerTaskData{brokerTaskData}
	} else {
		taskData, resp, err = restClient. BrokerTaskApi.ClustersClusterIdBrokersBrokerIdTasksGet(restContext, clusterId, brokerId)
	}
	if err != nil {
		return taskData, kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	return taskData, nil
}

func getBrokerTaskType(taskName string) (kafkarestv3.BrokerTaskType, error) {
	if taskName == "" {
		return "", nil
	}
	for _, taskType := range []kafkarestv3.BrokerTaskType{kafkarestv3.BROKERTASKTYPE_ADD_BROKER, kafkarestv3.BROKERTASKTYPE_REMOVE_BROKER} {
		if taskName == string(taskType) {
			return taskType, nil
		}
	}
	return "", errors.NewErrorWithSuggestions(errors.InvalidBrokerTaskTypeErrorMsg, errors.InvalidBrokerTaskTypeSuggestions)
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
