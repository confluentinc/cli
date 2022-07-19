package stream_share

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"
)

func (s *providerShareCommand) newListProviderShareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List provider shares.",
		Long:  "This command can be used to list shares for provider.",
		Args:  cobra.NoArgs,
		RunE:  s.listProviderShares,
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
