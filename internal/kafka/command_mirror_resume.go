package kafka

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
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
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
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

	kafkaREST, err := c.GetKafkaREST(cmd)
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
	isTruncateAndRestore := strings.HasPrefix(cmd.Use, "truncate-and-restore")
	includePartitionLevelTruncationData := false
	if isTruncateAndRestore {
		includePartitionData, err := cmd.Flags().GetBool(includePartitionDataFlagName)
		if err != nil {
			return err
		} else {
			includePartitionLevelTruncationData = includePartitionData
		}
	}
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

		var truncationData []*kafkarestv3.PartitionLevelTruncationData
		if includePartitionLevelTruncationData {
			nextPartitionDataIndex := 0
			for i := range result.GetMirrorLags().Items {
				if nextPartitionDataIndex >= len(result.GetPartitionLevelTruncationData().Items) {
					truncationData = append(truncationData, nil)
				} else {
					var data kafkarestv3.PartitionLevelTruncationData = result.GetPartitionLevelTruncationData().Items[nextPartitionDataIndex]
					if data.GetPartitionId() == int32(i) {
						truncationData = append(truncationData, &data)
						nextPartitionDataIndex++
					} else {
						truncationData = append(truncationData, nil)
					}
				}
			}
		}

		if !isTruncateAndRestore || includePartitionLevelTruncationData {
			for _, partitionLag := range result.GetMirrorLags().Items {
				partitionId := partitionLag.GetPartition()
				if isTruncateAndRestore && truncationData[partitionId] != nil {
					list.Add(&mirrorOut{
						MirrorTopicName:   result.GetMirrorTopicName(),
						Partition:         partitionId,
						ErrorMessage:      errorMessage,
						ErrorCode:         errorCode,
						MessagesTruncated: truncationData[partitionId].GetMessagesTruncated(),
						OffsetTruncatedTo: strconv.FormatInt(truncationData[partitionId].GetOffsetTruncatedTo(), 10),
					})
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
		} else {
			list.Add(&mirrorOut{
				MirrorTopicName:   result.GetMirrorTopicName(),
				ErrorMessage:      errorMessage,
				ErrorCode:         errorCode,
				MessagesTruncated: result.GetMessagesTruncated(),
			})
		}
	}
	if isTruncateAndRestore && includePartitionLevelTruncationData {
		list.Filter([]string{"MirrorTopicName", "Partition", "ErrorMessage", "ErrorCode", "OffsetTruncatedTo", "MessagesTruncated"})
	} else if isTruncateAndRestore {
		list.Filter([]string{"MirrorTopicName", "ErrorMessage", "ErrorCode", "MessagesTruncated"})
	} else {
		list.Filter([]string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode", "LastSourceFetchOffset"})
	}
	return list.Print()
}
