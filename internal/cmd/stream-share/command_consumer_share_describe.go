package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (s *consumerShareCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a consumer share.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(s.validArgs),
		RunE:              s.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe consumer share "ss-12345":`,
				Code: "confluent stream-share consumer share describe ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (s *consumerShareCommand) describe(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	consumerShare, _, err := s.V2Client.DescribeConsumerShare(shareId)
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, s.buildConsumerShare(consumerShare), consumerShareListFields, consumerHumanLabelMap, consumerStructuredLabelMap)
}
