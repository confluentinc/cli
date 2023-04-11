package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

type consumerGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type consumerData struct {
	ConsumerGroupId string `human:"Consumer Group" serialized:"consumer_group"`
	ConsumerId      string `human:"Consumer" serialized:"consumer"`
	InstanceId      string `human:"Instance" serialized:"instance"`
	ClientId        string `human:"Client" serialized:"client"`
}

type groupData struct {
	ClusterId         string
	ConsumerGroupId   string
	Coordinator       string
	IsSimple          bool
	PartitionAssignor string
	State             string
	Consumers         []consumerData
}

type consumerGroupOut struct {
	ClusterId         string `human:"Cluster" serialized:"cluster"`
	ConsumerGroupId   string `human:"Consumer Group" serialized:"consumer_group"`
	Coordinator       string `human:"Coordinator" serialized:"coordinator"`
	IsSimple          bool   `human:"Simple" serialized:"simple"`
	PartitionAssignor string `human:"Partition Assignor" serialized:"partition_assignor"`
	State             string `human:"State" serialized:"state"`
}

func newConsumerGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "consumer-group",
		Aliases:     []string{"cg"},
		Short:       "Manage Kafka consumer groups.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Hidden:      true,
	}

	c := &consumerGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

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
	consumerGroupDataList, err := listConsumerGroups(c.AuthenticatedCLICommand)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerGroupDataList.Data))
	for i, consumerGroup := range consumerGroupDataList.Data {
		suggestions[i] = consumerGroup.ConsumerGroupId
	}
	return suggestions
}

func listConsumerGroups(flagCmd *pcmd.AuthenticatedCLICommand) (*kafkarestv3.ConsumerGroupDataList, error) {
	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(flagCmd)
	if err != nil {
		return nil, err
	}

	groupCmdResp, httpResp, err := kafkaREST.CloudClient.ListKafkaConsumerGroups(lkc)
	if err != nil {
		return nil, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return &groupCmdResp, nil
}
