package cmd

import "github.com/spf13/cobra"

type HasAPIKeyCLICommand struct {
	*CLICommand
}

func NewHasAPIKeyCLICommand(cmd *cobra.Command, prerunner PreRunner) *HasAPIKeyCLICommand {
	c := &HasAPIKeyCLICommand{CLICommand: NewCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = prerunner.HasAPIKey(c)
	c.Command = cmd
	return c
}

func (h *HasAPIKeyCLICommand) AddCommand(cmd *cobra.Command) {
	cmd.PersistentPreRunE = h.PersistentPreRunE
	h.Command.AddCommand(cmd)
}
