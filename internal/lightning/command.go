package lightning

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lightning",
		Short: "Manage Lightning Tables.",
	}

	cmd.AddCommand(newQueryCommand(prerunner))

	return cmd
}
