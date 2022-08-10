package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type consumerShareCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newConsumerShareCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage consumer shares.",
	}

	c := &consumerShareCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}

func (s *consumerShareCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := s.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return s.autocompleteConsumerShares()
}

func (s *consumerShareCommand) autocompleteConsumerShares() []string {
	consumerShares, err := s.V2Client.ListConsumerShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerShares))
	for i, share := range consumerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}
