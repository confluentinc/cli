package kafka

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type brokerTaskData struct {
	ClusterId         string                     `human:"Cluster" serialized:"cluster_id"`
	BrokerId          int32                      `human:"Broker ID" serialized:"broker_id"`
	TaskType          kafkarestv3.BrokerTaskType `human:"Task Type" serialized:"task_type"`
	TaskStatus        string                     `human:"Task Status" serialized:"task_status"`
	CreatedAt         time.Time                  `human:"Created At" serialized:"created_at"`
	UpdatedAt         time.Time                  `human:"Updated At" serialized:"updated_at"`
	ShutdownScheduled bool                       `human:"Shutdown Scheduled,omitempty" serialized:"shutdown_scheduled,omitempty"`
	SubtaskStatuses   string                     `human:"Subtask Statuses" serialized:"subtask_statuses"`
	ErrorCode         int32                      `human:"Error Code,omitempty" serialized:"error_code,omitempty"`
	ErrorMessage      string                     `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
}

func (c *brokerCommand) newGetTasksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-tasks [id]",
		Short: "List broker tasks.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.getTasks,
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

	cmd.Flags().Bool("all", false, "List broker tasks for the cluster.")
	cmd.Flags().String("task-type", "", "Search by task type (add-broker or remove-broker).")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *brokerCommand) getTasks(cmd *cobra.Command, args []string) error {
	brokerId, all, err := broker.CheckAllOrIdSpecified(cmd, args, true)
	if err != nil {
		return err
	}

	taskType, err := cmd.Flags().GetString("task-type")
	if err != nil {
		return err
	}

	brokerTaskType, err := getBrokerTaskType(taskType)
	if err != nil {
		return err
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	var tasks []kafkarestv3.BrokerTaskData
	if all { // get BrokerTasks for the cluster
		tasks, err = getBrokerTasksForCluster(restClient, restContext, clusterId, brokerTaskType)
		if err != nil {
			return err
		}
	} else { // fetch individual broker configs
		tasks, err = getBrokerTasksForBroker(restClient, restContext, clusterId, brokerId, brokerTaskType)
		if err != nil {
			return err
		}
	}

	list := output.NewList(cmd)
	for _, entry := range tasks {
		list.Add(parseBrokerTaskData(entry))
	}
	return list.Print()
}

func parseBrokerTaskData(entry kafkarestv3.BrokerTaskData) *brokerTaskData {
	s := &brokerTaskData{
		ClusterId:       entry.ClusterId,
		BrokerId:        entry.BrokerId,
		TaskType:        entry.TaskType,
		TaskStatus:      entry.TaskStatus,
		CreatedAt:       entry.CreatedAt,
		UpdatedAt:       entry.UpdatedAt,
		SubtaskStatuses: mapToKeyValueString(entry.SubTaskStatuses),
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
		if kvString == "" {
			kvString = k + "=" + v
		} else {
			kvString = kvString + "\n" + k + "=" + v
		}
	}
	return kvString
}

func getBrokerTasksForCluster(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, taskType kafkarestv3.BrokerTaskType) ([]kafkarestv3.BrokerTaskData, error) {
	if taskType != "" {
		taskData, resp, err := restClient.BrokerTaskApi.ClustersClusterIdBrokersTasksTaskTypeGet(restContext, clusterId, taskType)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return taskData.Data, nil
	} else {
		taskData, resp, err := restClient.BrokerTaskApi.ClustersClusterIdBrokersTasksGet(restContext, clusterId)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return taskData.Data, nil
	}
}

func getBrokerTasksForBroker(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, brokerId int32, taskType kafkarestv3.BrokerTaskType) ([]kafkarestv3.BrokerTaskData, error) {
	if taskType != "" {
		tasks, resp, err := restClient.BrokerTaskApi.ClustersClusterIdBrokersBrokerIdTasksTaskTypeGet(restContext, clusterId, brokerId, taskType)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return []kafkarestv3.BrokerTaskData{tasks}, nil
	} else {
		tasks, resp, err := restClient.BrokerTaskApi.ClustersClusterIdBrokersBrokerIdTasksGet(restContext, clusterId, brokerId)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return tasks.Data, nil
	}
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
	return "", errors.NewErrorWithSuggestions(
		"invalid broker task type",
		`Valid broker task types are "remove-broker" and "add-broker".`,
	)
}
