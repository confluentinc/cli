package kafka

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var allowedMirrorTopicStatusValues = []string{"active", "failed", "paused", "stopped", "pending_stopped"}

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
				Code: "confluent kafka mirror list --link my-link --mirror-status ACTIVE",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String(mirrorStatusFlagName, "", fmt.Sprintf("Mirror topic status. Can be one of %s. If not specified, list all mirror topics.", utils.ArrayToCommaDelimitedString(allowedMirrorTopicStatusValues, "or")))
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
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	var mirrorTopicStatus *kafkarestv3.MirrorTopicStatus
	if mirrorStatus != "" {
		mirrorTopicStatus, err = kafkarestv3.NewMirrorTopicStatusFromValue(strings.ToUpper(mirrorStatus))
		if err != nil {
			return err
		}
	}

	var mirrors kafkarestv3.ListMirrorTopicsResponseDataList
	if linkName == "" {
		mirrors, err = kafkaREST.CloudClient.ListKafkaMirrorTopics(cluster.ID, mirrorTopicStatus)
		if err != nil {
			return err
		}
	} else {
		mirrors, err = kafkaREST.CloudClient.ListKafkaMirrorTopicsUnderLink(cluster.ID, linkName, mirrorTopicStatus)
		if err != nil {
			return err
		}
	}

	list := output.NewList(cmd)
	for _, mirror := range mirrors.GetData() {
		var maxLag int64 = 0
		for _, mirrorLag := range mirror.GetMirrorLags().Items {
			if lag := mirrorLag.GetLag(); lag > maxLag {
				maxLag = lag
			}
		}

		list.Add(&mirrorOut{
			LinkName:                 mirror.GetLinkName(),
			MirrorTopicName:          mirror.GetMirrorTopicName(),
			SourceTopicName:          mirror.GetSourceTopicName(),
			MirrorStatus:             string(mirror.GetMirrorStatus()),
			StatusTimeMs:             mirror.GetStateTimeMs(),
			NumPartition:             mirror.GetNumPartitions(),
			MaxPerPartitionMirrorLag: maxLag,
		})
	}
	list.Filter([]string{"LinkName", "MirrorTopicName", "NumPartition", "MaxPerPartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs"})
	return list.Print()
}
