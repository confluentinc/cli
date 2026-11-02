package logs

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func newSourceCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Discover the log sources you can query.",
		Long:  "Discover the log sources you can query with `confluent logs fetch`, including the search fields each one accepts and the fields it returns.",
	}

	cmd.AddCommand(newSourceDescribeCommand(prerunner))
	cmd.AddCommand(newSourceListCommand(prerunner))

	return cmd
}
