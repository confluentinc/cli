package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "byok",
		Short: "Manage your keys in Confluent Cloud.",
	}

	c := &command{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newRegisterCommand())
	c.AddCommand(c.newUnregisterCommand())

	return c.Command
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteEnvironments(c.Client, c.V2Client, c.Context)
}
