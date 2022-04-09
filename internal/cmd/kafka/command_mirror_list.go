package kafka

import (
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"net/http"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
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
				Text: `List all active mirror topics under "my-link":`,
				Code: "confluent kafka mirror list --link my-link --mirror-status active",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "Cluster link name. If not specified, list all mirror topics in the cluster.")
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

	mirrorStatusFlag, err := cmd.Flags().GetString(mirrorStatusFlagName)
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

	mirrorStatus := cloudkafkarest.MirrorTopicStatus(mirrorStatusFlag)

	var httpResp *http.Response
	var listMirrorTopicsResponse cloudkafkarest.ListMirrorTopicsResponseDataList

	if linkName == "" {
		req := kafkaREST.Client.ClusterLinkingV3Api.ListKafkaMirrorTopics(kafkaREST.Context, lkc)
		listMirrorTopicsResponse, httpResp, err = req.MirrorStatus(mirrorStatus).Execute()
	} else {
		req := kafkaREST.Client.ClusterLinkingV3Api.ListKafkaMirrorTopicsUnderLink(kafkaREST.Context, lkc, linkName)
		listMirrorTopicsResponse, httpResp, err = req.MirrorStatus(mirrorStatus).Execute()
	}
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listMirrorFields, humanListMirrorFields, structuredListMirrorFields)
	if err != nil {
		return err
	}

	for _, mirror := range listMirrorTopicsResponse.Data {
		var maxLag int64 = 0
		for _, mirrorLag := range mirror.MirrorLags.Items {
			if mirrorLag.Lag > maxLag {
				maxLag = mirrorLag.Lag
			}
		}

		outputWriter.AddElement(&listMirrorWrite{
			LinkName:                 mirror.LinkName,
			MirrorTopicName:          mirror.MirrorTopicName,
			SourceTopicName:          mirror.SourceTopicName,
			MirrorStatus:             string(mirror.MirrorStatus),
			StatusTimeMs:             mirror.StateTimeMs,
			NumPartition:             mirror.NumPartitions,
			MaxPerPartitionMirrorLag: maxLag,
		})
	}

	return outputWriter.Out()
}
