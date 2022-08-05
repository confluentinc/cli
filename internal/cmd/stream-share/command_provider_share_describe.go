package streamshare

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (s *providerShareCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a provider share.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(s.validArgs),
		RunE:              s.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe provider share "ss-12345":`,
				Code: "confluent stream-share provider share describe ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (s *providerShareCommand) describe(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	provideShare, _, err := s.V2Client.DescribeProvideShare(shareId)
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, s.buildProviderShare(provideShare), providerShareListFields, humanLabelMap, structuredLabelMap)
}
