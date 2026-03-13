package kafka

import (
	kafkarestv3Internal "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafkarest/v3"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupSubtopologyListCommand() *cobra.Command {
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

func (c *consumerCommand) listStreamGroupSubtopologies(cmd *cobra.Command, _ []string) error {
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

func (c *consumerCommand) getStreamGroupSubtopologies(cmd *cobra.Command) ([]kafkarestv3Internal.StreamsGroupSubtopologyData, error) {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := kafkaREST.CloudClientInternal.ListKafkaStreamsGroupMemberSubtopologies(groupId)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
