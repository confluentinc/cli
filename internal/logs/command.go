package logs

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Query logs from Confluent Cloud log sources.",
	}

	cmd.AddCommand(newFetchCommand(prerunner))
	cmd.AddCommand(newSourceCommand(prerunner))

	return cmd
}
