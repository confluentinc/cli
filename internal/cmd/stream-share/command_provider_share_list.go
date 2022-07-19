package stream_share

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"
)

func (s *providerShareCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shares for provider.",
		Args:  cobra.NoArgs,
		RunE:  s.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List provider shares",
				Code: "confluent stream-share provider share list",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}
