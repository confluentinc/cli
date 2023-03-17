package kafka

import (
	"fmt"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *mirrorCommand) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <destination-topic-1> [destination-topic-2] ... [destination-topic-N]",
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

	cobra.CheckErr(cmd.MarkFlagRequired(linkFlagName))

	return cmd
}

func (c *mirrorCommand) resume(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
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

	resumeMirrorOpt := &kafkarestv3.UpdateKafkaMirrorTopicsResumeOpts{
		AlterMirrorsRequestData: optional.NewInterface(kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: args}),
		ValidateOnly:            optional.NewBool(dryRun),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsResume(kafkaREST.Context, lkc, linkName, resumeMirrorOpt)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}

func printAlterMirrorResult(cmd *cobra.Command, results kafkarestv3.AlterMirrorStatusResponseDataList) error {
	list := output.NewList(cmd)
	for _, result := range results.Data {
		var errorMessage string
		if result.ErrorMessage != nil {
			errorMessage = *result.ErrorMessage
		}

		var errorCode string
		if result.ErrorCode != nil {
			errorCode = fmt.Sprint(*result.ErrorCode)
		}

		// fatal error
		if errorMessage != "" {
			list.Add(&mirrorOut{
				MirrorTopicName:       result.MirrorTopicName,
				Partition:             -1,
				ErrorMessage:          errorMessage,
				ErrorCode:             errorCode,
				PartitionMirrorLag:    -1,
				LastSourceFetchOffset: -1,
			})
			continue
		}

		for _, partitionLag := range result.MirrorLags {
			list.Add(&mirrorOut{
				MirrorTopicName:       result.MirrorTopicName,
				Partition:             partitionLag.Partition,
				ErrorMessage:          errorMessage,
				ErrorCode:             errorCode,
				PartitionMirrorLag:    partitionLag.Lag,
				LastSourceFetchOffset: partitionLag.LastSourceFetchOffset,
			})
		}
	}
	list.Filter([]string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode", "LastSourceFetchOffset"})
	return list.Print()
}
