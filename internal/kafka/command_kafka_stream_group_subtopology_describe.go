package kafka

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *consumerCommand) newStreamGroupSubtopologyDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <subtopology>",
		Short:             "Describe stream group subtopology",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStreamGroupArgs),
		RunE:              c.streamGroupSubtopologyDescribe,
	}

	cmd.Flags().String("group", "", "Group Id.")
	cmd.Flags().String("subtopology", "", "Subtopology Id.")

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))
	cobra.CheckErr(cmd.MarkFlagRequired("subtopology"))

	return cmd
}

func (c *consumerCommand) streamGroupSubtopologyDescribe(cmd *cobra.Command, args []string) error {
	groupId, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	subtopologyId, err := cmd.Flags().GetString("subtopology")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	subtopology, err := kafkaREST.CloudClientInternal.GetKafkaStreamGroupSubtopology(groupId, subtopologyId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&streamGroupSubtopologyOut{
		Kind:          subtopology.GetKind(),
		ClusterId:     subtopology.GetClusterId(),
		GroupId:       subtopology.GetGroupId(),
		SubtopologyId: subtopology.GetSubtopologyId(),
		SourceTopics:  subtopology.GetSourceTopics(),
	})

	return table.Print()
}
