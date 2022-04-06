package kafka

import (
	"fmt"
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *mirrorCommand) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <destination-topic-1> <destination-topic-2> ... <destination-topic-N>",
		Short: "Resume mirror topics.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.resume,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Resume mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror resume my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(linkFlagName)

	return cmd
}

func (c *mirrorCommand) resume(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetCloudKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	req := kafkaREST.Client.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsResume(kafkaREST.Context, lkc, linkName)
	req.AlterMirrorsRequestData(cloudkafkarest.AlterMirrorsRequestData{MirrorTopicNames: args}).ValidateOnly(validateOnly)
	results, httpResp, err := req.Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}
	return printAlterMirrorResult(cmd, results)
}

func printAlterMirrorResult(cmd *cobra.Command, results cloudkafkarest.AlterMirrorStatusResponseDataList) error {
	outputWriter, err := output.NewListOutputWriter(cmd, alterMirrorFields, humanAlterMirrorFields, structuredAlterMirrorFields)
	if err != nil {
		return err
	}

	for _, result := range results.Data {
		var errMsg = ""
		var code = ""

		if result.ErrorMessage.IsSet() {
			errMsg = *result.ErrorMessage.Get()
		}

		if result.ErrorCode.IsSet() {
			code = fmt.Sprint(*result.ErrorCode.Get())
		}

		// fatal error
		if errMsg != "" {
			outputWriter.AddElement(&alterMirrorWrite{
				MirrorTopicName:       result.MirrorTopicName,
				Partition:             -1,
				ErrorMessage:          errMsg,
				ErrorCode:             code,
				PartitionMirrorLag:    -1,
				LastSourceFetchOffset: -1,
			})
			continue
		}

		for _, partitionLag := range result.MirrorLags.Items {
			outputWriter.AddElement(&alterMirrorWrite{
				MirrorTopicName:       result.MirrorTopicName,
				Partition:             partitionLag.Partition,
				ErrorMessage:          errMsg,
				ErrorCode:             code,
				PartitionMirrorLag:    int64(partitionLag.Lag),
				LastSourceFetchOffset: partitionLag.LastSourceFetchOffset,
			})
		}
	}

	return outputWriter.Out()
}
