package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func newPrivateLinkCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "private-link",
		Short:   "Manage private links.",
		Aliases: []string{"pl"},
	}

	cmd.AddCommand(newPrivateLinkAccessCommand(prerunner))

	return cmd
}
