package stream_share

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"
)

func (s *providerShareCommand) newDescribeProviderShareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe provider share.",
		Args:  cobra.ExactArgs(1),
		RunE:  s.DescribeProviderShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe provider share",
				Code: "confluent stream-share provider share describe ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}
