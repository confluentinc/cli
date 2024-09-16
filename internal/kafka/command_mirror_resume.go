package kafka

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *mirrorCommand) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resume <destination-topic-1> [destination-topic-2] ... [destination-topic-N]",
		Short:             "Resume mirror topics.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.resume,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Resume mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror resume my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually resume the mirror topic, but simply validates it.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("link"))

	return cmd
}

func (c *mirrorCommand) resume(cmd *cobra.Command, args []string) error {
	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	alterMirrorsRequestData := kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: &args}

	results, err := kafkaREST.CloudClient.UpdateKafkaMirrorTopicsResume(link, dryRun, alterMirrorsRequestData)
	if err != nil {
		return err
	}

	return printAlterMirrorResult(cmd, results)
}

func printAlterMirrorResult(cmd *cobra.Command, results []kafkarestv3.AlterMirrorStatusResponseData) error {
	list := output.NewList(cmd)
	includingPartitionData := false
	for _, result := range results {
		errorMessage := result.GetErrorMessage()

		var errorCode string
		if code := result.GetErrorCode(); code != 0 {
			errorCode = strconv.Itoa(int(code))
		}

		// fatal error
		if errorMessage != "" {
			list.Add(&mirrorOut{
				MirrorTopicName:       result.GetMirrorTopicName(),
				Partition:             -1,
				ErrorMessage:          errorMessage,
				ErrorCode:             errorCode,
				PartitionMirrorLag:    -1,
				LastSourceFetchOffset: -1,
			})
			continue
		}

		var nextTruncationDataIndex = 0
		for _, partitionLag := range result.GetMirrorLags().Items {
			if len(result.GetPartitionLevelTruncationData().Items) > 0 && result.GetPartitionLevelTruncationData().Items[nextTruncationDataIndex].GetPartitionId() == partitionLag.GetPartition() {
				includingPartitionData = true
				messagesTruncated, _ := strconv.ParseInt(result.GetPartitionLevelTruncationData().Items[nextTruncationDataIndex].GetMessagesTruncated(), 10, 64)
				offsetTruncatedTo, _ := strconv.ParseInt(result.GetPartitionLevelTruncationData().Items[nextTruncationDataIndex].GetOffsetTruncatedTo(), 10, 64)
				list.Add(&mirrorOut{
					MirrorTopicName:       result.GetMirrorTopicName(),
					Partition:             partitionLag.GetPartition(),
					ErrorMessage:          errorMessage,
					ErrorCode:             errorCode,
					PartitionMirrorLag:    partitionLag.GetLag(),
					LastSourceFetchOffset: partitionLag.GetLastSourceFetchOffset(),
					MessagesTruncated:     messagesTruncated,
					OffsetTruncatedTo:     offsetTruncatedTo,
				})
				nextTruncationDataIndex += 1
			} else {
				list.Add(&mirrorOut{
					MirrorTopicName:       result.GetMirrorTopicName(),
					Partition:             partitionLag.GetPartition(),
					ErrorMessage:          errorMessage,
					ErrorCode:             errorCode,
					PartitionMirrorLag:    partitionLag.GetLag(),
					LastSourceFetchOffset: partitionLag.GetLastSourceFetchOffset(),
				})
			}
		}
	}
	if includingPartitionData {
		list.Filter([]string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode", "LastSourceFetchOffset", "OffsetTruncatedTo", "MessagesTruncated"})
	} else {
		list.Filter([]string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode", "LastSourceFetchOffset"})
	}
	err := list.Print()
	for _, result := range results {
		if result.GetMessagesTruncated() != "-1" {
			fmt.Println("Topic", result.GetMirrorTopicName(), "had a total of", result.GetMessagesTruncated(), "messages truncated.")
		}
	}
	return err
}
