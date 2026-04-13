package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *streamsGroupCommand) newStreamsGroupSubtopologyDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subtopology>",
		Short: "Describe a stream group subtopology.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.streamsGroupSubtopologyDescribe,
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

func (c *streamsGroupCommand) streamsGroupSubtopologyDescribe(cmd *cobra.Command, args []string) error {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	subtopologyId := args[0]

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	subtopology, err := kafkaREST.CloudClient.GetKafkaStreamsGroupSubtopology(groupId, subtopologyId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamsGroupSubtopologyOut{
		Kind:          subtopology.GetKind(),
		ClusterId:     subtopology.GetClusterId(),
		GroupId:       subtopology.GetGroupId(),
		SubtopologyId: subtopology.GetSubtopologyId(),
		SourceTopics:  subtopology.GetSourceTopics(),
	})

	return table.Print()
}
