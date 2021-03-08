package kafka

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"net/http"
)

// ahu: description should state 'max lag consumer ID', 'max lag instance ID', etc
var (
	lagSummaryHumanLabels = []string{"TotalLag", "MaxLag", "MaxLagConsumer", "MaxLagInstance", "MaxLagClient", "MaxLagTopic", "MaxLagPartition"}
	lagSummaryLabels = []string{"total_lag", "max_lag", "max_lag_consumer", "max_lag_instance", "max_lag_client", "max_lag_topic", "max_lag_partition"}



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

type consumerGroupLagSummaryData struct {
	TotalLag          int32  `json:"total_lag" yaml:"total_lag"`
	MaxLag            int32  `json:"max_lag" yaml:"max_lag"`
	MaxLagConsumerId  string `json:"max_lag_consumer" yaml:"max_lag_consumer"`
	MaxLagInstanceId  string `json:"max_lag_instance" yaml:"max_lag_instance"`
	MaxLagClientId    string `json:"max_lag_client" yaml:"max_lag_client"`
	MaxLagTopicName   string `json:"max_lag_topic" yaml:"max_lag_topic"`
	MaxLagPartitionId int32  `json:"max_lag_partition" yaml:"max_lag_partition"`
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
	lagCmd := NewLagCommand(c.prerunner)
	c.AddCommand(lagCmd.Command)
	c.completableChildren = lagCmd.completableChildren
}

func NewLagCommand(prerunner pcmd.PreRunner) *lagCommand {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:   "lag",
			Short: "View consumer lag.",
		}, prerunner, make(map[string]*pflag.FlagSet))
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
    lagCmd.AddCommand(summarizeLagCmd)

    listLagCmd := &cobra.Command{
    	Use:	"list <id>",
       	Short:	"List consumer lags for a Kafka consumer-group.",
      	Args:	cobra.ExactArgs(1),
       	RunE:	pcmd.NewCLIRunE(lagCmd.listLag),
       	Example: examples.BuildExampleString(
       		examples.Example{
       			Text: "List all consumer lags for consumers in consumer-group ``consumer-group-1``.",
       			Code: "ccloud kafka consumer-group lag list consumer-group-1",
       		},
       	),
    }
    lagCmd.AddCommand(listLagCmd)

   	getLagCmd := &cobra.Command{
    	Use:	"get <id>",
       	Short:	"Get consumer lag for a partition consumed by a Kafka consumer-group.",
      	Args:	cobra.ExactArgs(1),
       	RunE:	pcmd.NewCLIRunE(lagCmd.getLag),
       	Example: examples.BuildExampleString(
       		examples.Example{
       			Text: "Get the consumer lag for topic ``my_topic`` partition ``0`` consumed by consumer-group ``consumer-group-1``.",
       			Code: "ccloud kafka consumer-group lag get consumer-group-1 --topic my_topic --partition 0",
       		},
       	),
   	}
   	// ahu: handle defaults
   	getLagCmd.Flags().String("topic", "", "Topic name.")
   	getLagCmd.Flags().Int("partition", -1, "Partition ID.")
   	check(getLagCmd.MarkFlagRequired("topic"))
   	check(getLagCmd.MarkFlagRequired("partition"))
   	getLagCmd.Flags().SortFlags = false
   	lagCmd.AddCommand(getLagCmd)

   	lagCmd.completableChildren = []*cobra.Command{summarizeLagCmd, listLagCmd, getLagCmd}
}

func (lagCmd *lagCommand) summarizeLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	outputOption, err := lagCmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if !output.IsValidFormatString(outputOption) {
		return output.NewInvalidOutputFormatFlagError(outputOption)
	}

	kafkaREST, err := lagCmd.GetKafkaREST()
	if err != nil {
		return err
	}
	if kafkaREST != nil {
		kafkaClusterConfig, err := lagCmd.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID
		lagSummaryResp, httpResp, err :=
			kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGet(
				kafkaREST.Context,
				lkc,
				consumerGroupId)
		if httpResp != nil {
			// Kafka REST is available
			if err != nil {
				restErr, parseErr := parseOpenAPIError(err)
				if parseErr == nil && restErr.Code == KafkaRestUnknownConsumerGroupErrorCode {
					return fmt.Errorf(errors.UnknownGroupMsg, consumerGroupId)
				}
				// ahu: check if this will be descriptive enough to cover parse errors (if we can remove the preceding check)
				return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
			}
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST returns StatusOK
			consumerGroupLagSummaryData := &consumerGroupLagSummaryData{}
			consumerGroupLagSummaryData.TotalLag = lagSummaryResp.

		}


	}
	return err
}
