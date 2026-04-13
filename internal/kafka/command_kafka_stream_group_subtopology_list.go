package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamGroupCommand) newStreamGroupSubtopologyListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List kafka stream group subtopologies.",
		Args:        cobra.NoArgs,
		RunE:        c.listStreamGroupSubtopologies,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().String("group", "", "Group Id.")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *streamGroupCommand) listStreamGroupSubtopologies(cmd *cobra.Command, _ []string) error {
	subtopologies, err := c.getStreamGroupSubtopologies(cmd)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, subtopology := range subtopologies {
		list.Add(&streamGroupSubtopologyOut{
			Kind:          subtopology.GetKind(),
			ClusterId:     subtopology.GetClusterId(),
			GroupId:       subtopology.GetGroupId(),
			SubtopologyId: subtopology.GetSubtopologyId(),
			SourceTopics:  subtopology.GetSourceTopics(),
		})
	}

	return list.Print()
}

func (c *streamGroupCommand) getStreamGroupSubtopologies(cmd *cobra.Command) ([]kafkarestv3.StreamsGroupSubtopologyData, error) {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := kafkaREST.CloudClient.ListKafkaStreamsGroupMemberSubtopologies(groupId)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
