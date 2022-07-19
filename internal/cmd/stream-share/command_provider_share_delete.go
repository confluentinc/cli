package stream_share

import (
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"
)

func (s *providerShareCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a provider share.",
		Args:  cobra.ExactArgs(1),
		RunE:  s.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a provider share",
				Code: "confluent stream-share provider share delete ss-12345",
			},
		),
	}

	return cmd
}
