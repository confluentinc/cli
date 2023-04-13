package kafka

import (
	"net/http"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *mirrorCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mirror topics in a cluster or under a cluster link.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all mirror topics in the cluster:",
				Code: "confluent kafka mirror list --cluster lkc-1234",
			},
			examples.Example{
				Text: `List all active mirror topics under "my-link":`,
				Code: "confluent kafka mirror list --link my-link --mirror-status active",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String(mirrorStatusFlagName, "", "Mirror topic status. Can be one of [active, failed, paused, stopped, pending_stopped]. If not specified, list all mirror topics.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *mirrorCommand) list(cmd *cobra.Command, _ []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	mirrorStatus, err := cmd.Flags().GetString(mirrorStatusFlagName)
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

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	mirrorStatusOpt := optional.EmptyInterface()
	if mirrorStatus != "" {
		mirrorStatusOpt = optional.NewInterface(kafkarestv3.MirrorTopicStatus(mirrorStatus))
	}

	var listMirrorTopicsResponseDataList kafkarestv3.ListMirrorTopicsResponseDataList
	var httpResp *http.Response

	if linkName == "" {
		opts := &kafkarestv3.ListKafkaMirrorTopicsOpts{MirrorStatus: mirrorStatusOpt}
		listMirrorTopicsResponseDataList, httpResp, err = kafkaREST.Client.ClusterLinkingV3Api.ListKafkaMirrorTopics(kafkaREST.Context, lkc, opts)
	} else {
		opts := &kafkarestv3.ListKafkaMirrorTopicsUnderLinkOpts{MirrorStatus: mirrorStatusOpt}
		listMirrorTopicsResponseDataList, httpResp, err = kafkaREST.Client.ClusterLinkingV3Api.ListKafkaMirrorTopicsUnderLink(kafkaREST.Context, lkc, linkName, opts)
	}
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	list := output.NewList(cmd)
	for _, mirror := range listMirrorTopicsResponseDataList.Data {
		var maxLag int64 = 0
		for _, mirrorLag := range mirror.MirrorLags {
			if mirrorLag.Lag > maxLag {
				maxLag = mirrorLag.Lag
			}
		}

		list.Add(&mirrorOut{
			LinkName:                 mirror.LinkName,
			MirrorTopicName:          mirror.MirrorTopicName,
			SourceTopicName:          mirror.SourceTopicName,
			MirrorStatus:             string(mirror.MirrorStatus),
			StatusTimeMs:             mirror.StateTimeMs,
			NumPartition:             mirror.NumPartitions,
			MaxPerPartitionMirrorLag: maxLag,
		})
	}
	list.Filter([]string{"LinkName", "MirrorTopicName", "NumPartition", "MaxPerPartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs"})
	return list.Print()
}
