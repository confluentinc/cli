package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type providerShareCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newProviderShareCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage provider shares.",
	}

	c := &providerShareCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}

func (s *providerShareCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := s.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return s.autocompleteProviderShares()
}

func (s *providerShareCommand) autocompleteProviderShares() []string {
	providerShares, err := s.V2Client.ListProviderShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(providerShares))
	for i, share := range providerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}
