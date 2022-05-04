package kafka

import (
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var (
	lagFields               = []string{"ClusterId", "ConsumerGroupId", "Lag", "LogEndOffset", "CurrentOffset", "ConsumerId", "InstanceId", "ClientId", "TopicName", "PartitionId"}
	lagListHumanLabels      = []string{"Cluster", "ConsumerGroup", "Lag", "LogEndOffset", "CurrentOffset", "Consumer", "Instance", "Client", "Topic", "Partition"}
	lagListStructuredLabels = []string{"cluster", "consumer_group", "lag", "log_end_offset", "current_offset", "consumer", "instance", "client", "topic", "partition"}
	lagGetHumanRenames      = map[string]string{
		"ClusterId":       "Cluster",
		"ConsumerGroupId": "ConsumerGroup",
		"ConsumerId":      "Consumer",
		"InstanceId":      "Instance",
		"ClientId":        "Client",
		"TopicName":       "Topic",
		"PartitionId":     "Partition"}
	lagGetStructuredRenames = map[string]string{
		"ClusterId":       "cluster",
		"ConsumerGroupId": "consumer_group",
		"Lag":             "lag",
		"LogEndOffset":    "log_end_offset",
		"CurrentOffset":   "current_offset",
		"ConsumerId":      "consumer",
		"InstanceId":      "instance",
		"ClientId":        "client",
		"TopicName":       "topic",
		"PartitionId":     "partition"}
)

type consumerGroupCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type consumerData struct {
	ConsumerGroupId string `json:"consumer_group" yaml:"consumer_group"`
	ConsumerId      string `json:"consumer" yaml:"consumer"`
	InstanceId      string `json:"instance" yaml:"instance"`
	ClientId        string `json:"client" yaml:"client"`
}

type groupData struct {
	ClusterId         string         `json:"cluster" yaml:"cluster"`
	ConsumerGroupId   string         `json:"consumer_group" yaml:"consumer_group"`
	Coordinator       string         `json:"coordinator" yaml:"coordinator"`
	IsSimple          bool           `json:"simple" yaml:"simple"`
	PartitionAssignor string         `json:"partition_assignor" yaml:"partition_assignor"`
	State             string         `json:"state" yaml:"state"`
	Consumers         []consumerData `json:"consumers" yaml:"consumers"`
}

type groupDescribeStruct struct {
	ClusterId         string
	ConsumerGroupId   string
	Coordinator       string
	IsSimple          bool
	PartitionAssignor string
	State             string
}

type lagSummaryStruct struct {
	ClusterId         string
	ConsumerGroupId   string
	TotalLag          int64
	MaxLag            int64
	MaxLagConsumerId  string
	MaxLagInstanceId  string
	MaxLagClientId    string
	MaxLagTopicName   string
	MaxLagPartitionId int32
}

type lagDataStruct struct {
	ClusterId       string
	ConsumerGroupId string
	Lag             int64
	LogEndOffset    int64
	CurrentOffset   int64
	ConsumerId      string
	InstanceId      string
	ClientId        string
	TopicName       string
	PartitionId     int32
}

func newConsumerGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consumer-group",
		Aliases: []string{"cg"},
		Short:   "Manage Kafka consumer groups.",
		Hidden:  true,
	}

	c := &consumerGroupCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(newLagCommand(prerunner))
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func (c *consumerGroupCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConsumerGroups()
}

func (c *consumerGroupCommand) autocompleteConsumerGroups() []string {
	consumerGroupDataList, err := listConsumerGroups(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerGroupDataList.Data))
	for i, consumerGroup := range consumerGroupDataList.Data {
		suggestions[i] = consumerGroup.ConsumerGroupId
	}
	return suggestions
}

func listConsumerGroups(flagCmd *pcmd.AuthenticatedStateFlagCommand) (*kafkarestv3.ConsumerGroupDataList, error) {
	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(flagCmd)
	if err != nil {
		return nil, err
	}

	groupCmdResp, httpResp, err := kafkaREST.Client.ConsumerGroupV3Api.ListKafkaConsumerGroups(kafkaREST.Context, lkc)
	if err != nil {
		return nil, kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return &groupCmdResp, nil
}
