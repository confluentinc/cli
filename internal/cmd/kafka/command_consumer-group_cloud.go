package kafka

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

// ahu: description should state 'max lag consumer ID', 'max lag instance ID', etc
var (
	lagSummaryFields = []string{"ClusterId", "ConsumerGroupId", "TotalLag", "MaxLag", "MaxLagConsumerId", "MaxLagInstanceId", "MaxLagClientId", "MaxLagTopicName", "MaxLagPartitionId"}
	lagSummaryHumanRenames = map[string]string{
		"ClusterId":		 "Cluster",
		"ConsumerGroupId": 	 "ConsumerGroup",
		"MaxLagConsumerId":  "MaxLagConsumer",
		"MaxLagInstanceId":  "MaxLagInstance",
		"MaxLagClientId":    "MaxLagClient",
		"MaxLagTopicName":   "MaxLagTopic",
		"MaxLagPartitionId": "MaxLagPartition"}
	lagSummaryStructuredRenames = map[string]string{
		"ClusterId":		 "cluster",
		"ConsumerGroupId": 	 "consumer_group",
		"TotalLag":          "total_lag",
		"MaxLag":            "max_lag",
		"MaxLagConsumerId":  "max_lag_consumer",
		"MaxLagInstanceId":  "max_lag_instance",
		"MaxLagClientId":    "max_lag_client",
		"MaxLagTopicName":   "max_lag_topic",
		"MaxLagPartitionId": "max_lag_partition"}
)

type groupCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner			pcmd.PreRunner
	completableChildren []*cobra.Command
}

type lagCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner			pcmd.PreRunner
	completableChildren	[]*cobra.Command
}

type lagSummaryStruct struct {
	ClusterId 		  string
	ConsumerGroupId   string
	TotalLag          int32
	MaxLag            int32
	MaxLagConsumerId  string
	MaxLagInstanceId  string
	MaxLagClientId    string
	MaxLagTopicName   string
	MaxLagPartitionId int32
}

func NewGroupCommand(prerunner pcmd.PreRunner) *groupCommand {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:	"consumer-group",
			Short:	"Manage Kafka consumer-groups.",
		}, prerunner, GroupSubcommandFlags)
	groupCommand := &groupCommand{
		AuthenticatedStateFlagCommand:	cliCmd,
		prerunner:						prerunner,
	}
	groupCommand.init()
	return groupCommand
}

func (c *groupCommand) init() {
	//listCmd := &cobra.Command{
	//	Use:   "list",
	//	Short: "List Kafka consumer groups.",
	//	Args:  cobra.NoArgs,
	//	RunE:  pcmd.NewCLIRunE(c.list),
	//}
	//listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	//listCmd.Flags().SortFlags = false
	//c.AddCommand(listCmd)
	//
	//describeCmd := &cobra.Command{
	//	Use:   "describe <consumer-group>",
	//	Short: "Describe a Kafka consumer group.",
	//	Args:  cobra.ExactArgs(1),
	//	RunE:  pcmd.NewCLIRunE(c.describe),
	//}
	//describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	//describeCmd.Flags().SortFlags = false
	//c.AddCommand(describeCmd)

	lagCmd := NewLagCommand(c.prerunner)
	c.AddCommand(lagCmd.Command)

	//c.completableChildren = append(lagCmd.completableChildren, listCmd)
	c.completableChildren = lagCmd.completableChildren
}

func NewLagCommand(prerunner pcmd.PreRunner) *lagCommand {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:   "lag",
			Short: "View consumer lag.",
		}, prerunner, LagSubcommandFlags)
	lagCmd := &lagCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		prerunner:                     prerunner,
	}
	lagCmd.init()
	return lagCmd
}

func (lagCmd *lagCommand) init() {
    summarizeLagCmd := &cobra.Command{
    	Use:	"summarize <id>",
    	Short:	"Summarize consumer lag for a Kafka consumer-group.",
    	Args:	cobra.ExactArgs(1),
    	RunE:	pcmd.NewCLIRunE(lagCmd.summarizeLag),
    	Example: examples.BuildExampleString(
    		examples.Example{
    			Text: "Summarize the lag for consumer-group ``consumer-group-1``.",
    			// ahu: should the examples include the other required flag(s)? --cluster
    			Code: "ccloud kafka consumer-group lag summarize consumer-group-1",
    		},
    	),
    }
    summarizeLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
    summarizeLagCmd.Flags().SortFlags = false
    lagCmd.AddCommand(summarizeLagCmd)

    //listLagCmd := &cobra.Command{
    //	Use:	"list <id>",
    //   	Short:	"List consumer lags for a Kafka consumer-group.",
    //  	Args:	cobra.ExactArgs(1),
    //   	RunE:	pcmd.NewCLIRunE(lagCmd.listLag),
    //   	Example: examples.BuildExampleString(
    //   		examples.Example{
    //   			Text: "List all consumer lags for consumers in consumer-group ``consumer-group-1``.",
    //   			Code: "ccloud kafka consumer-group lag list consumer-group-1",
    //   		},
    //   	),
    //}
    //lagCmd.AddCommand(listLagCmd)
	//
   	//getLagCmd := &cobra.Command{
    //	Use:	"get <id>",
    //   	Short:	"Get consumer lag for a partition consumed by a Kafka consumer-group.",
    //  	Args:	cobra.ExactArgs(1),
    //   	RunE:	pcmd.NewCLIRunE(lagCmd.getLag),
    //   	Example: examples.BuildExampleString(
    //   		examples.Example{
    //   			Text: "Get the consumer lag for topic ``my_topic`` partition ``0`` consumed by consumer-group ``consumer-group-1``.",
    //   			Code: "ccloud kafka consumer-group lag get consumer-group-1 --topic my_topic --partition 0",
    //   		},
    //   	),
   	//}
   	//// ahu: handle defaults
   	//getLagCmd.Flags().String("topic", "", "Topic name.")
   	//getLagCmd.Flags().Int("partition", -1, "Partition ID.")
   	//check(getLagCmd.MarkFlagRequired("topic"))
   	//check(getLagCmd.MarkFlagRequired("partition"))
   	//getLagCmd.Flags().SortFlags = false
   	//lagCmd.AddCommand(getLagCmd)

   	//lagCmd.completableChildren = []*cobra.Command{summarizeLagCmd, listLagCmd, getLagCmd}
	lagCmd.completableChildren = []*cobra.Command{summarizeLagCmd}
}

func (lagCmd *lagCommand) summarizeLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, err := lagCmd.GetKafkaREST()
	if err != nil {
		return err
	}
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailable)
	}
	// Kafka REST is available
	kafkaClusterConfig, err := lagCmd.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	lkc := kafkaClusterConfig.ID
	fmt.Print("got the lkc ")
	fmt.Println(lkc)
	lagSummaryResp, _, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGet(
			kafkaREST.Context,
			lkc,
			consumerGroupId)

	if err != nil {
		return err
	}

	return output.DescribeObject(
		cmd,
		convertLagSummaryToStruct(lagSummaryResp),
		lagSummaryFields,
		lagSummaryHumanRenames,
		lagSummaryStructuredRenames)

	//lagSummaryResp, httpResp, err :=
	//	kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGet(
	//		kafkaREST.Context,
	//		lkc,
	//		consumerGroupId)
	//
	//if httpResp != nil {
	//	fmt.Print("httpResp received ")
	//	if err != nil {
	//		fmt.Print("error getting lag response ")
	//		restErr, parseErr := parseOpenAPIError(err)
	//		if parseErr == nil && restErr.Code == KafkaRestUnknownConsumerGroupErrorCode {
	//			return fmt.Errorf(errors.UnknownGroupMsg, consumerGroupId)
	//		}
	//		// ahu: check if this will be descriptive enough to cover parse errors (if we can remove the preceding check)
	//		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	//	}
	//	if httpResp.StatusCode != http.StatusOK {
	//		fmt.Print("got a status code that wasn't OK ")
	//		return errors.NewErrorWithSuggestions(
	//			fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
	//			errors.InternalServerErrorSuggestions)
	//	}
	//	// Kafka REST returns StatusOK
	//	fmt.Print("we got status OK ")
	//	return output.DescribeObject(
	//		cmd,
	//		convertLagSummaryToStruct(lagSummaryResp),
	//		lagSummaryFields,
	//		lagSummaryHumanRenames,
	//		lagSummaryStructuredRenames)
	//}
	//fmt.Print("no httpResp received ")
	//return err

}

func convertLagSummaryToStruct(lagSummaryData kafkarestv3.ConsumerGroupLagSummaryData) *lagSummaryStruct {
	maxLagInstanceId := ""
	if lagSummaryData.MaxLagInstanceId != nil {
		maxLagInstanceId = *lagSummaryData.MaxLagInstanceId
	}
	return &lagSummaryStruct{
		ClusterId:		   lagSummaryData.ClusterId,
		ConsumerGroupId:   lagSummaryData.ConsumerGroupId,
		TotalLag:          lagSummaryData.TotalLag,
		MaxLag:            lagSummaryData.MaxLag,
		MaxLagConsumerId:  lagSummaryData.MaxLagConsumerId,
		MaxLagInstanceId:  maxLagInstanceId,
		MaxLagClientId:    lagSummaryData.MaxLagClientId,
		MaxLagTopicName:   lagSummaryData.MaxLagTopicName,
		MaxLagPartitionId: lagSummaryData.MaxLagPartitionId,
	}
}

func (c *groupCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *groupCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	return suggestions
}

func (c *groupCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}
